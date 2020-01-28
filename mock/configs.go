package mock

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ConfigsRequest struct {
	ReleaseKey string  `form:"releaseKey"`
	Ip         string  `form:"ip"`
	Messages   Message `form:"messages"`
}

type ConfigsResponse struct {
	AppId          string            `json:"appId"`
	Cluster        string            `json:"cluster"`
	NamespaceName  string            `json:"namespaceName"`
	ReleaseKey     string            `json:"releaseKey"`
	Configurations map[string]string `json:"configurations"`
}

func (s *Server) handleConfigs(c *gin.Context) {
	appId := c.Param("appId")
	cluster := c.Param("cluster")
	namespaceName := c.Param("namespaceName")

	req := ConfigsRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusForbidden, err.Error())
		fmt.Printf("ShouldBindQuery fail:%s", err.Error())
		return
	}
	fmt.Printf("appId:%s\n", appId)
	fmt.Printf("cluster:%s\n", cluster)
	fmt.Printf("namespaceName:%s\n", namespaceName)
	fmt.Printf("req:%+v\n", req)

	// if key's length is less than 10, then return not modified response
	if len(req.ReleaseKey) < 10 {
		c.JSON(http.StatusNotModified, nil)
		return
	}

	configurations := make(map[string]string)
	configurations["a1"] = req.ReleaseKey
	configurations["a2"] = req.ReleaseKey

	releaseKey := req.ReleaseKey
	if len(releaseKey) == 0 {
		releaseKey = "releaseKey1"
	} else {
		releaseKey = releaseKey + "z"
	}

	resp := ConfigsResponse{
		AppId:          appId,
		Cluster:        cluster,
		NamespaceName:  namespaceName,
		ReleaseKey:     releaseKey,
		Configurations: configurations,
	}

	c.JSON(http.StatusOK, resp)
	return
}
