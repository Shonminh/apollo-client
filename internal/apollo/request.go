// Forked from https://github.com/zouyx/agollo

package apollo

import (
	"github.com/Shonminh/apollo-client/internal/logger"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type CallBack struct {
	SuccCallBack      func([]byte) (interface{}, error)
	NotModifyCallBack func() error
}

type reqConfig struct {
	// 设置到http.client中timeout字段
	Timeout time.Duration
	// 连接接口的uri
	Uri string
}

func requestRecovery(hostRs Resolver, rc *reqConfig, callBack *CallBack) (interface{}, error) {
	var retErr = NewMutliError()
	var err error
	var response interface{}

	hosts, err := hostRs.resolve()
	if err != nil {
		return nil, errors.WithMessage(err, "resolve")
	}
	for _, v := range hosts {
		requestUrl := v + "/" + rc.Uri
		response, err = request(requestUrl, rc.Timeout, callBack)
		if err != nil {
			logger.LogInfo("request faield, %v, %v", requestUrl, err)
			retErr = multierror.Append(retErr, err)
			continue
		}
		return response, nil
	}

	return nil, errors.WithMessage(retErr, "all hosts failed")
}

func request(requestUrl string, timeout time.Duration, callBack *CallBack) (interface{}, error) {
	client := &http.Client{Timeout: ConnectTimeout}
	// 如有设置自定义超时时间即使用
	if timeout != 0 {
		client.Timeout = timeout
	}

	var resBody []byte
	var retErr = NewMutliError()
	var res *http.Response
	var waitTs time.Duration = RetryInterval
	for i := 0; i < max_retries; i++ {
		var err error
		if i > 0 {
			time.Sleep(waitTs)
			waitTs += RetryInterval
		}
		res, err = client.Get(requestUrl)
		if res == nil || err != nil {
			logger.LogError("Connect Apollo Server Fail,Error: %v waitTs %s", err, waitTs)
			retErr = multierror.Append(retErr, err)
			continue
		}

		// not modified break
		switch res.StatusCode {
		case http.StatusOK:
			resBody, err = ioutil.ReadAll(res.Body)
			if err != nil {
				logger.LogError("Connect Apollo Server Fail,Error: %v", err)
				retErr = multierror.Append(retErr, err)
				continue
			}

			if callBack != nil && callBack.SuccCallBack != nil {
				return callBack.SuccCallBack(resBody)
			}
			return nil, nil
		case http.StatusNotModified:
			logger.LogInfo("Config Not Modified")
			if callBack != nil && callBack.NotModifyCallBack != nil {
				return nil, callBack.NotModifyCallBack()
			}
			return nil, nil
		case http.StatusGatewayTimeout:
			logger.LogInfo("Gateway Timeout")
			return nil, nil
		default:
			logger.LogError("Connect Apollo Server Fail,StatusCode: %v", res.StatusCode)
			err = errors.WithMessage(ErrInvalidHttpStatus, "status "+strconv.Itoa(res.StatusCode))
			retErr = multierror.Append(retErr, err)
			continue
		}
	}

	if retErr.ErrorOrNil() != nil {
		logger.LogError("failed after max retry %v %v", max_retries, retErr)
		return nil, retErr
	}
	return nil, nil
}
