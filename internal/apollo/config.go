// Forked from https://github.com/zouyx/agollo

package apollo

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"time"
)

var (
	max_retries    = 0
	ConnectTimeout = 5 * time.Second
	RetryInterval  = 5 * time.Second
	NofityTimeout  = 65 * time.Second
)

type (
	ConnConfig struct {
		AppId         string `json:"appId"`
		Cluster       string `json:"cluster"`
		NamespaceName string `json:"namespaceName"`
		ReleaseKey    string `json:"releaseKey"`
	}

	Config struct {
		ConnConfig
		Configurations map[string]string `json:"configurations"`
	}

	ConfigCenter struct {
		Host     Resolver
		AppId    string `json:"appId"`
		Cluster  string `json:"cluster"`
		ClientIp string `json:"-"`
	}

	Message struct {
		Details map[string]interface{} `json:"details"`
	}
)

func (cc *ConfigCenter) getMessages(namespaceName string, notificationId int64) (string, error) {
	key := fmt.Sprintf("%s+%s+%s", cc.AppId, cc.Cluster, namespaceName)
	message := Message{}
	message.Details = make(map[string]interface{})
	message.Details[key] = notificationId

	result, err := json.Marshal(message)
	if err != nil {
		return "", errors.WithMessage(err, "json.Marshal")
	}
	return string(result), nil
}

func (cc *ConfigCenter) SyncConfig(namespaceName, releaseKey string, notificationId int64) (ret *Config, err error) {
	message, err := cc.getMessages(namespaceName, notificationId)
	if err != nil {
		err = errors.WithMessage(err, "getMessages")
		return
	}

	urlSuffix := cc.getConfigUrlSuffix(namespaceName, releaseKey, message)

	cf, err := requestRecovery(
		cc.Host,
		&reqConfig{
			Uri: urlSuffix,
		},
		&CallBack{
			SuccCallBack: syncSucc,
		},
	)
	if err != nil {
		err = errors.WithMessage(err, "requestRecovery")
	}

	if cf != nil {
		ret = cf.(*Config)
	}

	if ret != nil {
		if ret.ReleaseKey == releaseKey {
			err = errors.New("same config compare with last")
		}

		if ret.NamespaceName != namespaceName {
			err = errors.New(fmt.Sprintf("namespace miss match: [%v, %v]", ret.NamespaceName, namespaceName))
		}
	}

	return
}

func (cc *ConfigCenter) getConfigUrlSuffix(namespaceName, releaseKey, message string) string {
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s&messages=%s",
		url.QueryEscape(cc.AppId),
		url.QueryEscape(cc.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(releaseKey),
		cc.ClientIp,
		url.QueryEscape(message))
}

func syncSucc(resBody []byte) (interface{}, error) {
	conf, err := jsonToConfig(resBody)
	if err != nil {
		return conf, errors.WithMessage(err, "jsonToConfig failed")
	}
	return conf, nil
}

func jsonToConfig(b []byte) (*Config, error) {
	ret := &Config{}
	err := json.Unmarshal(b, ret)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal failed")
	}
	return ret, nil
}
