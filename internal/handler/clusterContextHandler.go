package handler

import (
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/zkerrors"
)

var LOG_TAG2 = "ClusterContextHandler"

type ClusterConfigHandler struct {
	OpLogin  *auth.OperatorLogin
	ZkConfig *config.ZkOperatorConfig
}

type ClusterContextResponse struct {
	ClusterKey string `json:"clusterKey"`
	CloudAddr  string `json:"cloudAddr"`
	ClusterId  string `json:"clusterId"`
}

func (h *ClusterConfigHandler) ClusterContextHandler(ctx iris.Context) {

	response := ClusterContextResponse{}
	response.ClusterId = h.OpLogin.GetClusterId()
	logger.Debug(LOG_TAG, h.ZkConfig.ZkCloud)
	response.CloudAddr = h.ZkConfig.ZkCloud.Host + ":" + h.ZkConfig.ZkCloud.Port

	clusterKey, err := utils.GetSecretValue(h.ZkConfig.OperatorLogin.ClusterKeyNamespace, h.ZkConfig.OperatorLogin.ClusterKey, h.ZkConfig.OperatorLogin.ClusterKeyData)

	var zkError *zkerrors.ZkError
	if err != nil {
		logger.Error(LOG_TAG2, " Cluster Context api ", err.Error())
		zkErrorTemp := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		zkError = &zkErrorTemp
		response.ClusterKey = ""
	} else {
		response.ClusterKey = clusterKey
	}

	zkHttpResponse := zkhttp.ToZkResponse[ClusterContextResponse](200, response, response, zkError)

	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)

}

func (h *ClusterConfigHandler) CleanUpOnkill() error {
	logger.Debug(LOG_TAG2, "Nothing to clean here.")
	return nil
}
