package handler

import (
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/auth"
	logger "github.com/zerok-ai/zk-utils-go/logs"
)

var LOG_TAG2 = "ClusterContextHandler"

type ClusterConfigHandler struct {
	OpLogin *auth.OperatorLogin
}

func (h *ClusterConfigHandler) ClusterContextHandler(ctx iris.Context) {
	clusterId := h.OpLogin.GetClusterId()
	logger.Debug(LOG_TAG2, clusterId)
}

func (h *ClusterConfigHandler) CleanUpOnkill() error {
	logger.Debug(LOG_TAG2, "Nothing to clean here.")
	return nil
}
