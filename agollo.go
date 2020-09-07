// Forked from https://github.com/zouyx/agollo

package agollo

import (
	"encoding/json"
	"fmt"
	"github.com/Shonminh/apollo-client/internal/apollo"
	"github.com/Shonminh/apollo-client/internal/logger"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"go.uber.org/ratelimit"
	"golang.org/x/net/context"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	// ConfigReader gets type-specified value
	ConfigReader interface {
		GetStringValue(key string) (string, error)
		GetIntValue(key string) (int, error)
		GetFloatValue(key string) (float64, error)
		GetBoolValue(key string) (bool, error)
		GetBytesValue(key string) ([]byte, error)
	}
	configReader string

	namespace struct {
		NamespaceName  string `json:"namespaceName"`
		releaseKey     string `json:"-"`
		NotificationId int64  `json:"notificationId"`
	}
	service struct {
		apollo.ConfigCenter
		namespaceList []*namespace
	}

	defaultVal struct {
		s string
		i int
		f float64
		b bool
	}
)

// CHandler calls a handler to process config change event
type CHandler func(event *ChangeEvent) error

const (
	DEFAULT_NOFICATION_ID = -1
)

var (
	gOption    *option
	gInitOnce  sync.Once
	gStartOnce sync.Once
	gService   *service
	gDefault   map[string]defaultVal
	gCallback  CHandler
	rateLimit  = ratelimit.New(2)
)

func Init(opts ...Option) (err error) {
	gInitOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
				return
			}
		}()
		gOption = newDefaultOption()
		for _, opt := range opts {
			opt.apply(gOption)
		}

		if gOption.defaultVals != nil {
			gDefault = gOption.defaultVals
		}

		if gOption.ApolloAddr == "" {
			err = errors.New("ApolloAddr not set")
			return
		}
		initCache(gOption.ConfigCacheSize, gOption.ignoreNameSpace)

		apollo.NofityTimeout = gOption.notifyTimeout
		apollo.ConnectTimeout = gOption.connectTimeout
		apollo.RetryInterval = gOption.retryInterval
		s := &service{
			ConfigCenter: apollo.ConfigCenter{
				Host:     apollo.NewSingleHostResolver(gOption.ApolloAddr),
				AppId:    gOption.AppId,
				Cluster:  gOption.Cluster,
				ClientIp: gOption.clientIp,
			},
		}
		nl := strings.Split(gOption.NamespaceName, ",")
		s.namespaceList = make([]*namespace, len(nl))
		for i, v := range nl {
			s.namespaceList[i] = &namespace{
				NamespaceName:  v,
				releaseKey:     "",
				NotificationId: DEFAULT_NOFICATION_ID,
			}
		}
		gService = s
		if !gOption.quickInitWithBK {
			err = s.syncConfig(true, nil)
			if err != nil {
				err = errors.WithMessage(err, "syncConfig init")
				return
			}
		} else {
			err = s.LoadConfigFile(nil)
			if err != nil {
				err = errors.WithMessage(err, "syncConfig init")
				return
			}
		}
	})
	return err
}

// Start starts agollo
func Start() {
	StartContext(context.Background())
}

// StartContext starts agollo
func StartContext(ctx context.Context) {
	gStartOnce.Do(func() {
		gService.start(ctx)
	})
}

// GetConfigReader returns a reader that can read config values
func GetConfigReader(namespaceName string) ConfigReader {
	return configReader(namespaceName)
}

// start agollo
func (s *service) start(ctx context.Context) {
	t1 := time.NewTimer(gOption.refreshInterval)
	for {
		select {
		case <-t1.C:
			_ = s.syncConfig(false, nil)
			t1.Reset(gOption.refreshInterval)
		case <-ctx.Done():
			break
		default:
			rateLimit.Take()
			s.pullNotify()
		}
	}
}

func (s *service) syncConfig(isInit bool, nm []*namespace) error {
	if nm == nil {
		nm = s.namespaceList
	}
	var retErr *multierror.Error
	var event = make(map[string]*ChangeEvent)
	for _, v := range nm {
		cfg, err := s.ConfigCenter.SyncConfig(v.NamespaceName, v.releaseKey, v.NotificationId)
		if err != nil || (cfg == nil && isInit) {
			logger.LogError(fmt.Sprintf("sync namespace [%s] config failed %v", v.NamespaceName, err))
			retErr = multierror.Append(apollo.NewMutliError(), err)
			if isInit {
				path := getConfigPath(v.NamespaceName)
				cfg, err = loadConfigFile(path)
				if err != nil {
					retErr = multierror.Append(retErr, errors.WithMessage(err, "loadConfigFile "+v.NamespaceName))
					continue
				}
			}
		}
		if cfg != nil {
			event[v.NamespaceName] = getChangeEvent(cfg)
			updateCache(cfg, v, event[v.NamespaceName])
		}
	}
	if !isInit {
		for _, v := range nm {
			if e, ok := event[v.NamespaceName]; ok {
				pushChange(v, e)
			}
		}
	}
	// err only print log, dont return, cause maybe namespace success once in namespace list.
	if retErr != nil {
		logger.LogError(fmt.Sprintf("syncConfig failed %s", retErr.Error()))
	}
	return nil
}

func (s *service) LoadConfigFile(nm []*namespace) error {
	if nm == nil {
		nm = s.namespaceList
	}
	for _, v := range nm {
		path := getConfigPath(v.NamespaceName)
		cfg, err := loadConfigFile(path)
		if err != nil {
			return errors.WithMessage(err, "loadConfigFile "+v.NamespaceName)
		}

		if cfg != nil {
			if cfg.NamespaceName != v.NamespaceName {
				return errors.New(fmt.Sprintf("namespace miss match: [%v, %v]", cfg.NamespaceName, v.NamespaceName))
			}
			event := getChangeEvent(cfg)
			updateCache(cfg, v, event)
		}
	}
	return nil
}

func (s *service) pullNotify() {
	nfs, err := s.getNotifies()
	if err != nil {
		logger.LogError("error getNotifies: %v", err)
		return
	}
	nf, err := s.ConfigCenter.PullNotify(nfs)
	if err != nil {
		logger.LogError("error PullNotify: %v", err)
		return
	}
	// nothing changed
	if len(nf) == 0 {
		return
	}

	nmList := s.updateNotificationId(nf)
	s.syncConfig(false, nmList)
}

func (s *service) getNotifies() (string, error) {
	j, err := json.Marshal(s.namespaceList)
	if err != nil {
		return "", errors.WithMessage(err, "json.Marshal")
	}
	return string(j), nil
}

func (s *service) updateNotificationId(nf []*apollo.NotifyMsg) (updated []*namespace) {
	// 关注的 namespace 不会太多
	for _, v := range nf {
		for _, v2 := range s.namespaceList {
			if v.NamespaceName == v2.NamespaceName {
				v2.NotificationId = v.NotificationId
				updated = append(updated, v2)
				break
			}
		}
	}
	return updated
}

func getConfigPath(namespaceName string) string {
	return gOption.BackupDir + "/" + namespaceName + gOption.BackupSuffix
}

func (cr configReader) getValue(key string) (string, error) {
	ck := getCacheKey(string(cr), key)
	cacheMutex.Lock()
	value, err := gConfigCache.Get([]byte(ck))
	cacheMutex.Unlock()
	if err != nil {
		return EMPTY, errors.WithMessage(err, "getValue "+key)
	}

	return string(value), nil
}

func (cr configReader) GetStringValue(key string) (string, error) {
	value, err := cr.getValue(key)
	if err != nil {
		def := gDefault[getCacheKey(string(cr), key)]
		return def.s, errors.WithMessage(err, "getValue")
	}
	return value, nil
}

func (cr configReader) GetIntValue(key string) (int, error) {
	value, err := cr.getValue(key)
	if err != nil {
		def := gDefault[getCacheKey(string(cr), key)]
		return def.i, errors.WithMessage(err, "getValue")
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		def := gDefault[getCacheKey(string(cr), key)]
		return def.i, errors.WithMessage(err, "atoi")
	}
	return parsed, nil
}

func (cr configReader) GetFloatValue(key string) (float64, error) {
	value, err := cr.getValue(key)
	if err != nil {
		def := gDefault[getCacheKey(string(cr), key)]
		return def.f, errors.WithMessage(err, "getValue")
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		def := gDefault[getCacheKey(string(cr), key)]
		return def.f, errors.WithMessage(err, "ParseFloat")
	}

	return parsed, nil
}

func (cr configReader) GetBoolValue(key string) (bool, error) {
	value, err := cr.getValue(key)
	if err != nil {
		def := gDefault[getCacheKey(string(cr), key)]
		return def.b, errors.WithMessage(err, "getValue")
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		def := gDefault[getCacheKey(string(cr), key)]
		return def.b, errors.WithMessage(err, "ParseBool")
	}

	return parsed, nil
}

func (cr configReader) GetBytesValue(key string) ([]byte, error) {
	ck := getCacheKey(string(cr), key)
	cacheMutex.Lock()
	value, err := gConfigCache.Get([]byte(ck))
	cacheMutex.Unlock()
	if err != nil {
		return nil, errors.WithMessage(err, "GetBytesValue")
	}
	return value, nil
}

// RegChangeEventHandler register a config change handler on agollo
func RegChangeEventHandler(in CHandler) {
	gCallback = in
}

// get namespace list being watched
func GetNamespaceList() []string {
	var ret []string
	if gService != nil {
		ret = make([]string, len(gService.namespaceList))
		for k, v := range gService.namespaceList {
			ret[k] = v.NamespaceName
		}
	}
	return ret
}
