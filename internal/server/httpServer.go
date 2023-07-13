package server

import (
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/handler"
	logger "github.com/zerok-ai/zk-utils-go/logs"
)

var LOG_TAG_HTTP = "HttpServer"

func exceptionHandler(ctx iris.Context) {
	ctx.StatusCode(iris.StatusOK)
}

func StartHttpServer(app *iris.Application, config iris.Configurator, zkConfig config.ZkOperatorConfig,
	clusterContextHandler *handler.ClusterContextHandler) {

	httpServerConfig := zkConfig.Http

	app.Post(httpServerConfig.ExceptionPath, exceptionHandler)

	logger.Debug(LOG_TAG_HTTP, zkConfig.ClusterContext.Path)

	app.Get(zkConfig.ClusterContext.Path, clusterContextHandler.Handler)

	err := app.Run(iris.Addr(":"+httpServerConfig.Port), config)
	if err != nil {
		logger.Error(LOG_TAG_HTTP, "Error while starting http server ", err)
		return
	}

}
