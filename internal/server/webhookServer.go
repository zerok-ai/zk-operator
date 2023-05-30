package server

import (
	"bytes"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/inject"
	"github.com/zerok-ai/zk-operator/internal/storage"
	"io"

	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/config"
)

type WebhookRequestHandler struct {
	injector *inject.Injector
}

func (h *WebhookRequestHandler) ServeHTTP(ctx iris.Context) {
	body, err := io.ReadAll(ctx.Request().Body)

	fmt.Printf("Got a request from webhook")

	if err != nil {
		webhookErrorResponse(err, ctx, "Failed to ready body of webhook request.")
		return
	}

	response, err := h.injector.Inject(body)

	if err != nil {
		fmt.Printf("Error while injecting zk agent %v\n", err)
	}

	// Sending http status as OK, even when injection failed to not disturb the pods in cluster.
	ctx.StatusCode(iris.StatusOK)
	ctx.Write(response)
}

func webhookErrorResponse(err error, ctx iris.Context, message string) {
	fmt.Printf("%v with error %v.\n", message, err)
	ctx.StatusCode(iris.StatusInternalServerError)
}

func handleRoutes(app *iris.Application, cfg config.ZkInjectorConfig, runtimeMap *storage.ImageRuntimeCache) {
	injectHandler := &WebhookRequestHandler{
		injector: &inject.Injector{ImageRuntimeHandler: runtimeMap, Config: cfg},
	}
	app.Post(cfg.Webhook.Path, injectHandler.ServeHTTP)
}

func StartWebHookServer(app *iris.Application, cfg config.ZkInjectorConfig, cert *bytes.Buffer, key *bytes.Buffer, runtimeMap *storage.ImageRuntimeCache, config iris.Configurator) {
	handleRoutes(app, cfg, runtimeMap)
	app.Run(iris.TLS(":"+cfg.Webhook.Port, cert.String(), key.String()), config)
}
