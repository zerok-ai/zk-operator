package handler

import (
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/auth"
	logger "github.com/zerok-ai/zk-utils-go/logs"
)

var LOG_TAG2 = "ClusterContextHandler"

var opLogin *auth.OperatorLogin

func SetOpLogin(opLoginParam *auth.OperatorLogin) {
	opLogin = opLoginParam
}

func ClusterContextHandler(ctx iris.Context) {
	clusterId := opLogin.GetClusterId()
	logger.Debug(LOG_TAG2, clusterId)
}
