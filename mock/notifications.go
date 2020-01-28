package mock

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type NotificationsRequest struct {
	AppId         string `form:"appId" binding:"required"`
	Cluster       string `form:"cluster" binding:"required"`
	Notifications string `form:"notifications" binding:"required"`
}

type Notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationId int64  `json:"notificationId"`
}

type NotificationRes struct {
	NamespaceName  string  `json:"namespaceName"`
	NotificationId int64   `json:"notificationId"`
	Messages       Message `json:"messages"`
}

type Message struct {
	Details map[string]interface{} `json:"details"`
}

func (s *Server) handleNotifications(c *gin.Context) {
	req := NotificationsRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusForbidden, err.Error())
		fmt.Printf("ShouldBindQuery fail:%s", err.Error())
		return
	}
	fmt.Printf("req:%+v\n", req)

	var resp []NotificationRes

	var notifications []Notification
	err := json.Unmarshal([]byte(req.Notifications), &notifications)
	if err != nil {
		c.JSON(http.StatusForbidden, err.Error())
		return
	}

	for _, notification := range notifications {
		// if notificationId is lower than 50, then nothing change
		if notification.NotificationId < 50 {
			continue
		}

		key := fmt.Sprintf("%s+%s+%s", req.AppId, req.Cluster, notification.NamespaceName)
		message := Message{}
		message.Details = make(map[string]interface{})
		notificationId := notification.NotificationId + 1
		message.Details[key] = notificationId

		resItem := NotificationRes{
			NamespaceName:  notification.NamespaceName,
			NotificationId: notificationId,
			Messages:       message,
		}

		resp = append(resp, resItem)
	}

	if len(resp) == 0 {
		c.JSON(http.StatusNotModified, nil)
		return
	}

	c.JSON(http.StatusOK, resp)
	return
}
