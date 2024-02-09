package server

import (
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/handler"
	logger "github.com/zerok-ai/zk-utils-go/logs"
)

var LOG_TAG_HTTP = "HttpServer"

func StartHttpServer(app *iris.Application, config iris.Configurator, zkConfig config.ZkOperatorConfig, modules []internal.ZkOperatorModule) {

	httpServerConfig := zkConfig.Http
	logger.Debug(LOG_TAG_HTTP, zkConfig.ClusterContext.Path)

	healthCheckHandler := handler.HealthCheckHandler{}
	healthCheckHandler.Init(modules)

	app.Get("/healthz", healthCheckHandler.Handler)

	err := app.Run(iris.Addr(":"+httpServerConfig.Port), config)
	if err != nil {
		logger.Error(LOG_TAG_HTTP, "Error while starting http server ", err)
		return
	}
}
