package handler

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/config"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
)

var LOG_TAG2 = "ClusterContextHandler"

type ClusterContextHandler struct {
	ZkConfig  *config.ZkOperatorConfig
	ClusterId string
}

func (h *ClusterContextHandler) IsHealthy() bool {
	return true
}

type ClusterContextResponse struct {
	CloudAddr string `json:"cloudAddr"`
	ClusterId string `json:"clusterId"`
}

func (h *ClusterContextHandler) Handler(ctx iris.Context) {

	response := ClusterContextResponse{}
	response.ClusterId = h.ClusterId
	logger.Debug(scenarioLogTag, h.ZkConfig.ZkCloud)
	addr := fmt.Sprintf("%v:%v", h.ZkConfig.ClusterContext.CloudAddr, h.ZkConfig.ClusterContext.Port)
	response.CloudAddr = addr

	zkHttpResponse := zkhttp.ToZkResponse[ClusterContextResponse](200, response, response, nil)

	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)

}

func (h *ClusterContextHandler) CleanUpOnKill() error {
	logger.Debug(LOG_TAG2, "Nothing to clean here.")
	return nil
}
