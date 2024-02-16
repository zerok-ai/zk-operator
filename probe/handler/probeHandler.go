package handler

import (
	"errors"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/probe/model/response"
	"github.com/zerok-ai/zk-operator/probe/service"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"gopkg.in/yaml.v2"
	"io"
	"reflect"
)

type ProbeHandler interface {
	GetAllProbes(ctx iris.Context)
	DeleteProbe(ctx iris.Context)
	UpdateProbe(ctx iris.Context)
	CreateProbe(ctx iris.Context)
	GetAllServices(ctx iris.Context)
}

const LogTag = "probeHandler"

type probeHandler struct {
	service service.ProbeService
	cfg     config.ZkOperatorConfig
}

func NewProbeHandler(service service.ProbeService) ProbeHandler {
	zklogger.Error(LogTag, "NewProbeHandler******....******")
	return &probeHandler{service: service}
}

func (p *probeHandler) GetAllProbes(ctx iris.Context) {
	zklogger.Error(LogTag, "GetAllProbes************")
	resp, zkErr := p.service.GetAllProbes()
	fmt.Println(LogTag, resp)
	fmt.Println(LogTag, zkErr)
	zkHttpResponse := zkhttp.ToZkResponse[response.CRDListResponse](200, resp, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	fmt.Println(reflect.TypeOf(zkHttpResponse))
	err := ctx.JSON(zkHttpResponse)
	if err != nil {
		zklogger.Error(LogTag, "Error marshalling response", err)
		return
	}

}

func (p *probeHandler) DeleteProbe(ctx iris.Context) {
	zklogger.Error(LogTag, "DeleteProbe************")
	zkErr := p.service.DeleteProbe(ctx.Params().Get("name"))
	zkHttpResponse := zkhttp.ToZkResponse[any](200, nil, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func (p *probeHandler) UpdateProbe(ctx iris.Context) {
	zklogger.Error(LogTag, "UpdateProbe************")
	probeBody, err := readProbeRequest(ctx)
	if err != nil {
		zklogger.Error(LogTag, "Error reading probe request", err)
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Error reading probe request"})
		return
	}

	err = validateProbeBody(probeBody)
	if err != nil {
		zklogger.Error(LogTag, "Error validating probe body", err)
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Error validating probe body"})
		return
	}

	zkErr := p.service.UpdateProbe(probeBody)
	zkHttpResponse := zkhttp.ToZkResponse[any](200, nil, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func (p *probeHandler) CreateProbe(ctx iris.Context) {
	zklogger.Error(LogTag, "CreateProbe************")
	probeBody, err := readProbeRequest(ctx)
	if err != nil {
		zklogger.Error(LogTag, "Error reading probe request", err)
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Error reading probe request"})
		return
	}

	err = validateProbeBody(probeBody)
	if err != nil {
		zklogger.Error(LogTag, "Error validating probe body", err)
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Error validating probe body"})
		return
	}

	zkErr := p.service.CreateProbe(probeBody)
	zkHttpResponse := zkhttp.ToZkResponse[any](200, nil, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func (p *probeHandler) GetAllServices(ctx iris.Context) {
	zklogger.Error(LogTag, "GetAllServices************")
	resp, zkErr := p.service.GetAllServices()
	zkHttpResponse := zkhttp.ToZkResponse[response.ServiceListResponse](200, resp, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func readProbeRequest(ctx iris.Context) (v1alpha1.ZerokProbeSpec, error) {
	var req v1alpha1.ZerokProbeSpec
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return req, err
	}

	err = yaml.Unmarshal(body, req)
	if err != nil {
		return req, err
	}

	return req, nil
}

func validateProbeBody(probeBody v1alpha1.ZerokProbeSpec) error {
	if probeBody.Title == "" {
		zklogger.Error(LogTag, "Title cannot be empty")
		return errors.New("title cannot be empty")
	}

	if probeBody.Workloads == nil {
		zklogger.Error(LogTag, "Workloads cannot be empty")
		return errors.New("workloads cannot be empty")
	}

	if probeBody.Filter.Condition == "" {
		zklogger.Error(LogTag, "Filter cannot be empty")
		return errors.New("filter cannot be empty")
	}

	probeBody.GroupBy = nil
	probeBody.RateLimit = nil
	return nil
}
