package handler

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/zkerrors"
)

var LOG_TAG2 = "ClusterContextHandler"

type ClusterContextHandler struct {
	OpLogin  *auth.OperatorLogin
	ZkConfig *config.ZkOperatorConfig
}

func (h *ClusterContextHandler) IsHealthy() bool {
	return true
}

type ClusterContextResponse struct {
	ApiKey    string `json:"apiKey"`
	CloudAddr string `json:"cloudAddr"`
	ClusterId string `json:"clusterId"`
}

func (h *ClusterContextHandler) Handler(ctx iris.Context) {

	response := ClusterContextResponse{}
	response.ClusterId = h.OpLogin.GetClusterId()
	logger.Debug(scenarioLogTag, h.ZkConfig.ZkCloud)
	addr := fmt.Sprintf("%v:%v", h.ZkConfig.ClusterContext.CloudAddr, h.ZkConfig.ClusterContext.Port)
	response.CloudAddr = addr

	apiKey, err := utils.GetSecretValue(h.ZkConfig.OperatorLogin.ClusterKeyNamespace, h.ZkConfig.OperatorLogin.ClusterSecretName, h.ZkConfig.OperatorLogin.ApiKeyData)

	var zkError *zkerrors.ZkError
	if err != nil {
		logger.Error(LOG_TAG2, " Cluster Context api ", err.Error())
		zkErrorTemp := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		zkError = &zkErrorTemp
		response.ApiKey = ""
	} else {
		response.ApiKey = apiKey
	}

	zkHttpResponse := zkhttp.ToZkResponse[ClusterContextResponse](200, response, response, zkError)

	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)

}

func (h *ClusterContextHandler) CleanUpOnKill() error {
	logger.Debug(LOG_TAG2, "Nothing to clean here.")
	return nil
}
