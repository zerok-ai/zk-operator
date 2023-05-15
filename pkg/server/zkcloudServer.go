package server

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/zerok-ai/zk-operator/internal/config"
	zkhttp "github.com/zerok-ai/zk-operator/pkg/common/zkhttp"
	"github.com/zerok-ai/zk-operator/pkg/server/models"
	"github.com/zerok-ai/zk-operator/pkg/zkclient"

	"github.com/kataras/iris/v12"
)

type ZkCloudApiHandler struct {
}

func (h *ZkCloudApiHandler) ServeHTTP(ctx iris.Context) {
	body, err := io.ReadAll(ctx.Request().Body)

	fmt.Printf("Got a request from zk cloud.\n")

	if err != nil {

		zkCloudErrorResponse(err, ctx)
		return
	}

	err = h.restartWorkloads(body)

	//TODO: Do we have any format for sending api response body.?
	if err != nil {
		zkCloudErrorResponse(err, ctx)
	} else {
		ctx.StatusCode(iris.StatusOK)
	}
}

func (h *ZkCloudApiHandler) restartWorkloads(body []byte) error {
	restartRequestObj := models.RestartRequest{}
	if err := json.Unmarshal(body, &restartRequestObj); err != nil {
		return fmt.Errorf("unmarshaling restart request failed with %s", err)
	}
	namespace := restartRequestObj.Namespace
	all := restartRequestObj.All
	if all {
		return zkclient.RestartAllDeploymentsInNamespace(namespace)
	} else {
		deployment := restartRequestObj.Deployment
		return zkclient.RestartDeployment(namespace, deployment)
	}
}

func zkCloudErrorResponse(err error, ctx iris.Context) {
	fmt.Printf("Error while restarting workloads %v\n", err)
	zkReponse := zkhttp.ZkHttpResponseBuilder[any]{}.WithZkErrorType(zkhttp.ZK_ERROR_INTERNAL_SERVER).Build()
	ctx.StatusCode(zkReponse.Status)
	ctx.JSON(zkReponse)
}

func handleZkCloudRoutes(app *iris.Application, cfg config.ZkInjectorConfig) {
	apiHandler := &ZkCloudApiHandler{}
	//Adding new route for zk cloud.
	fmt.Println("Adding new route for zk cloud.")
	app.Post(cfg.ZkCloud.RestartPath, apiHandler.ServeHTTP)
}

func StartZkCloudServer(app *iris.Application, cfg config.ZkInjectorConfig, irisConfig iris.Configurator) {
	handleZkCloudRoutes(app, cfg)
	app.Run(iris.Addr(":"+cfg.ZkCloud.Port), irisConfig)
}
