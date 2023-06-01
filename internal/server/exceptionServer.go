package server

import (
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/config"
)

func dummyHandler(ctx iris.Context) {
	ctx.StatusCode(iris.StatusOK)
}

func StartExceptionServer(app *iris.Application, config iris.Configurator, exceptionConfig config.ExceptionConfig) {
	app.Post(exceptionConfig.Path, dummyHandler)
	app.Run(iris.Addr(":"+exceptionConfig.Port), config)
}
