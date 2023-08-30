package server

import (
	"bytes"
	"github.com/zerok-ai/zk-operator/internal/inject"
	"github.com/zerok-ai/zk-operator/internal/storage"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"io"

	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/config"
)

var LOG_TAG string = "WebhookServer"

type WebhookRequestHandler struct {
	injector *inject.Injector
}

func (h *WebhookRequestHandler) ServeHTTP(ctx iris.Context) {
	body, err := io.ReadAll(ctx.Request().Body)

	logger.Info(LOG_TAG, "Got a request from webhook")

	if err != nil {
		webhookErrorResponse(err, ctx, "Failed to ready body of webhook request.")
		return
	}

	response, err := h.injector.Inject(body)

	if err != nil {
		logger.Error(LOG_TAG, "Error while injecting zk agent ", err)
	}

	// Sending http status as OK, even when injection failed to not disturb the pods in cluster.
	ctx.StatusCode(iris.StatusOK)
	ctx.Write(response)
}

func webhookErrorResponse(err error, ctx iris.Context, message string) {
	logger.Error(LOG_TAG, message, " with error ", err)
	ctx.StatusCode(iris.StatusInternalServerError)
}

func handleRoutes(app *iris.Application, cfg config.ZkOperatorConfig, runtimeMap *storage.ImageRuntimeCache, appInitContainerData *config.AppInitContainerData) {
	injectHandler := &WebhookRequestHandler{
		injector: &inject.Injector{ImageRuntimeHandler: runtimeMap, Config: cfg, InitContainerData: appInitContainerData},
	}
	app.Post(cfg.Webhook.Path, injectHandler.ServeHTTP)
}

func StartWebHookServer(app *iris.Application, cfg config.ZkOperatorConfig, cert *bytes.Buffer, key *bytes.Buffer, runtimeMap *storage.ImageRuntimeCache, config iris.Configurator, appInitContainerData *config.AppInitContainerData) {
	logger.Debug(LOG_TAG, "Starting webhook server.")
	handleRoutes(app, cfg, runtimeMap, appInitContainerData)
	err := app.Run(iris.TLS(":"+cfg.Webhook.Port, cert.String(), key.String()), config)
	if err != nil {
		logger.Error(LOG_TAG, "Error while starting webhook server ", err)
		return
	}
}
