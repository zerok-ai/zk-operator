package handler

import (
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/utils"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
)

var healthCheckTag = "healthCheckHandler"

type HealthCheckHandler struct {
	ZkModules []internal.ZkOperatorModule
}

func (h *HealthCheckHandler) Init(zkModules []internal.ZkOperatorModule) {
	h.ZkModules = zkModules
}

func (h *HealthCheckHandler) Handler(ctx iris.Context) {
	if len(h.ZkModules) == 0 {
		ctx.StatusCode(502)
	}
	isHealthy := true
	for _, module := range h.ZkModules {
		if !module.IsHealthy() {
			moduleName := utils.GetTypeName(module)
			zklogger.Debug(healthCheckTag, "Module ", moduleName, " is not healthy.")
			isHealthy = false
			break
		}
	}
	if isHealthy {
		ctx.StatusCode(200)
	} else {
		ctx.StatusCode(500)
	}
}
