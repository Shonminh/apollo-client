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
	// 默认agollo配置文件名
	DEFAULT_CONFFILE = "app.properties"
	// 默认备份文件目录
	DEFAULT_BACKUPDIR = "."
	// 默认备份文件后缀
	DEFAULT_BACKUPSUFFIX = ".apollo.json"
	// 默认缓存大小
	DEFAULT_CONFIGCACHESIZE = 50 * 1024 * 1024
	// 默认 namespace
	DEFAULT_NAMESPACENAME = "application"
	// 默认 cluster
	DEFAULT_CLUSTER = "default"

	// 默认 notify 超时时间, 参见 apollo notifications 接口
	DEFAULT_NOTIFYTIMEOUT = 65 * time.Second

	// 默认连接超时时间
	DEFAULT_CONNECTTIMEOUT = 10 * time.Second
	// 出错后重试的暂停时间
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

//Option contains apply method to load options
type Option interface {
	apply(*option)
}

// WithDefaultVals 传一个map, 用于指定配置项的默认值, 如
//    WithDefaultVals(map[string]interface{}{
//		"a": "11",
//	  }, "application")
//    设定了 application 配置项 a 的默认值为 "11"
//    当调用 GetStringValue("a") 时，如果 a 配置项不存在，返回 11
//    注意: 默认值限定类型 GetIntValue("a") 不会默认返回 11，还是默认返回 0
//    注意: 默认值必须是 int/string/bool/float64, 否则会崩溃
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

// 设置定时配置刷新间隔
func WithRefreshInterval(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.refreshInterval = v
	})
}

// 设置实时配置推断更新间隔
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

// 设置要拉取的namespace, 可以传多个，逗号间隔
func WithNamespaceName(s string) Option {
	return newFuncOption(func(o *option) {
		o.NamespaceName = s
	})
}

// 设置apollo配置中心地址, 如:
//    127.0.0.1:8888
//    https://meta.apollo.com
func WithApolloAddr(s string) Option {
	return newFuncOption(func(o *option) {
		o.ApolloAddr = s
	})
}

// 设置备份文件存放目录
func WithBackupDir(s string) Option {
	return newFuncOption(func(o *option) {
		o.BackupDir = s
	})
}

// 设置备份文件后缀名
func WithBackupSuffix(s string) Option {
	return newFuncOption(func(o *option) {
		o.BackupSuffix = s
	})
}

// 设置agollo配置文件
func WithConfFile(s string) Option {
	return newFuncOption(func(o *option) {
		o.confFile = s
		err := loadJsonConfig(gOption, gOption.confFile)
		if err != nil {
			panic(errors.WithMessage(err, "loadJsonConfig"))
		}
	})
}

// 设置缓存大小
func WithCacheSize(v int) Option {
	return newFuncOption(func(o *option) {
		o.ConfigCacheSize = v
	})
}

// 设置client ip, apollo 会记录这个ip
func WithClientIp(v string) Option {
	return newFuncOption(func(o *option) {
		o.clientIp = v
	})
}

// 设置 apollo notify 接口超时时间
func WithNotifyTimeout(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.notifyTimeout = v
	})
}

// 设置 apollo configs 接口超时时间
func WithConnectTimeout(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.connectTimeout = v
	})
}

// 设置连接出错后重试的间隔时间
func WithRetryInterval(v time.Duration) Option {
	return newFuncOption(func(o *option) {
		o.retryInterval = v
	})
}

// Init 时, 从本地备份文件中加载初始配置(不连接apollo)
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
