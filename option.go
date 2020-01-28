package agollo

import (
	"encoding/json"
	"fmt"
	"github.com/Shonminh/apollo-client/internal/logger"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

const (
	// default apollo config file
	DEFAULT_CONFFILE = "app.properties"
	// default backup directory
	DEFAULT_BACKUPDIR = "."
	// default backup file suffix
	DEFAULT_BACKUPSUFFIX = ".apollo.json"
	// default cache size
	DEFAULT_CONFIGCACHESIZE = 50 * 1024 * 1024
	// default namespace
	DEFAULT_NAMESPACENAME = "application"
	// default cluster
	DEFAULT_CLUSTER = "default"

	// default notify timeout, see apollo notifications
	DEFAULT_NOTIFYTIMEOUT = 65 * time.Second

	// default connect timeout
	DEFAULT_CONNECTTIMEOUT = 10 * time.Second
	// default retry interval
	DEFAULT_RETRYINTERVAL = 3 * time.Second
)

var (
	DEFAULT_REFRESHINTERVAL  = 5 * time.Minute
	DEFAULT_LONGPOLLINTERVAL = 5 * time.Second
)

type option struct {
	AppId           string `json:"appId"`
	Cluster         string `json:"cluster"`
	NamespaceName   string `json:"namespaceName"`
	ApolloAddr      string `json:"apolloAddr"`
	BackupDir       string `json:"backupDir"`
	BackupSuffix    string `json:"backupSuffix"`
	ConfigCacheSize int    `json:"configCacheSize"`

	confFile         string
	refreshInterval  time.Duration
	longPollInterval time.Duration

	defaultVals map[string]defaultVal

	clientIp string

	notifyTimeout  time.Duration
	connectTimeout time.Duration
	retryInterval  time.Duration

	quickInitWithBK bool
}

func newDefaultOption() *option {
	return &option{
		refreshInterval:  DEFAULT_REFRESHINTERVAL,
		longPollInterval: DEFAULT_LONGPOLLINTERVAL,
		confFile:         DEFAULT_CONFFILE,
		BackupDir:        DEFAULT_BACKUPDIR,
		BackupSuffix:     DEFAULT_BACKUPSUFFIX,
		ConfigCacheSize:  DEFAULT_CONFIGCACHESIZE,

		NamespaceName: DEFAULT_NAMESPACENAME,
		Cluster:       DEFAULT_CLUSTER,

		clientIp: getInternal(),

		notifyTimeout:  DEFAULT_NOTIFYTIMEOUT,
		connectTimeout: DEFAULT_CONNECTTIMEOUT,
		retryInterval:  DEFAULT_RETRYINTERVAL,
	}
}

// Option contains apply method to load options
type Option interface {
	apply(*option)
}

// set key's default value
// for example:
//    WithDefaultVals(map[string]interface{}{
//		"key1": "11",
//	  }, "application")
//    when call GetStringValue("key1"), if config key1 not found, return 11
//   NOTE: default value is bound to type, if call GetIntValue("key1") will return 0
//  NOTE: default value's type is in int/string/bool/float64, will panic when use other types
func WithDefaultVals(val map[string]interface{}, namespaceName string) Option {
	return newFuncOption(func(o *option) {
		var err error
		o.defaultVals = make(map[string]defaultVal)
		for k, v := range val {
			ck := getCacheKey(namespaceName, k)
			o.defaultVals[ck], err = newDefaultVal(v)
			if err != nil {
				panic(errors.WithMessage(err, "default value for key "+k))
			}
		}
	})
}

// set refresh interval
func WithRefreshInterval(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.refreshInterval = v
	})
}

// set long poll interval
func WithLongPollInterval(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.longPollInterval = v
	})
}

func WithAppId(s string) Option {
	return newFuncOption(func(o *option) {
		o.AppId = s
	})
}

func WithCluster(s string) Option {
	return newFuncOption(func(o *option) {
		o.Cluster = s
	})
}

// set namespace list, use , to separate each namespace
func WithNamespaceName(s string) Option {
	return newFuncOption(func(o *option) {
		o.NamespaceName = s
	})
}

// set apollo address, for example:
//    127.0.0.1:8888
//    https://meta.apollo.com
func WithApolloAddr(s string) Option {
	return newFuncOption(func(o *option) {
		o.ApolloAddr = s
	})
}

// set backup save directory
func WithBackupDir(s string) Option {
	return newFuncOption(func(o *option) {
		o.BackupDir = s
	})
}

// set backup file's suffix name
func WithBackupSuffix(s string) Option {
	return newFuncOption(func(o *option) {
		o.BackupSuffix = s
	})
}

// set apollo config file
func WithConfFile(s string) Option {
	return newFuncOption(func(o *option) {
		o.confFile = s
		err := loadJsonConfig(gOption, gOption.confFile)
		if err != nil {
			panic(errors.WithMessage(err, "loadJsonConfig"))
		}
	})
}

// set cache size
func WithCacheSize(v int) Option {
	return newFuncOption(func(o *option) {
		o.ConfigCacheSize = v
	})
}

// set client ip
func WithClientIp(v string) Option {
	return newFuncOption(func(o *option) {
		o.clientIp = v
	})
}

// set apollo notify timeout
func WithNotifyTimeout(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.notifyTimeout = v
	})
}

// set apollo config connect timeout
func WithConnectTimeout(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.connectTimeout = v
	})
}

// set retry interval
func WithRetryInterval(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.retryInterval = v
	})
}

// get config from backup file when call Init function
func WithQuickInitWithBK() Option {
	return newFuncOption(func(o *option) {
		o.quickInitWithBK = true
	})
}

func WithLogFunc(logDebug, logInfo, logError logger.LogFunc) Option {
	return newFuncOption(func(o *option) {
		logger.LogDebug = logDebug
		logger.LogInfo = logInfo
		logger.LogError = logError
	})
}


type funcOption struct {
	f func(*option)
}

func (fo *funcOption) apply(o *option) {
	fo.f(o)
}

func newFuncOption(f func(*option)) *funcOption {
	return &funcOption{f: f}
}

func loadJsonConfig(opt *option, fileName string) error {
	fs, err := ioutil.ReadFile(fileName)
	if err != nil {
		return errors.WithMessage(err, "ReadFile")
	}

	err = json.Unmarshal(fs, opt)
	if err != nil {
		return errors.WithMessage(err, "json.Unmarshal")
	}

	return nil
}

func newDefaultVal(i interface{}) (defaultVal, error) {
	ret := defaultVal{}
	switch i.(type) {
	case int:
		ret.i = i.(int)
	case float64:
		ret.f = i.(float64)
	case bool:
		ret.b = i.(bool)
	case string:
		ret.s = i.(string)
	default:
		msg := fmt.Sprintf("val %v is not int/float64/bool/string", i)
		return ret, errors.New(msg)
	}
	return ret, nil
}
