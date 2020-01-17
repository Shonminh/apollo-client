package apollo

import (
	"encoding/json"
	"fmt"
	"github.com/Shonminh/apollo-client/internal/logger"
	"github.com/pkg/errors"
	"net/url"
)

type (
	NotifyMsg struct {
		NotificationId int64  `json:"notificationId"`
		NamespaceName  string `json:"namespaceName"`
	}
)

func (cc *ConfigCenter) PullNotify(notifications string) ([]*NotifyMsg, error) {
	urlSuffix := getNotifyUrlSuffix(notifications, cc.AppId, cc.Cluster)

	notifies, err := requestRecovery(
		cc.Host,
		&reqConfig{
			Uri:     urlSuffix,
			Timeout: NofityTimeout,
		},
		&CallBack{
			SuccCallBack: notifySucc,
		},
	)
	if err != nil {
		err = errors.WithMessage(err, "requestRecovery")
	}

	if notifies == nil {
		return nil, err
	}

	return notifies.([]*NotifyMsg), err
}

func getNotifyUrlSuffix(notifications, appId, cluster string) string {
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(appId),
		url.QueryEscape(cluster),
		url.QueryEscape(notifications))
}

func notifySucc(resBody []byte) (o interface{}, err error) {
	return toNotify(resBody)
}

func toNotify(resBody []byte) ([]*NotifyMsg, error) {
	ret := make([]*NotifyMsg, 0)

	err := json.Unmarshal(resBody, &ret)

	if err != nil {
		logger.LogError("Unmarshal Msg Fail,Error: %v", err)
		return nil, err
	}
	return ret, nil
}
