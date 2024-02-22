package handler

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/probe/model/request"
	"github.com/zerok-ai/zk-operator/probe/model/response"
	"github.com/zerok-ai/zk-operator/probe/service"
	"github.com/zerok-ai/zk-utils-go/common"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"gopkg.in/yaml.v2"
)

type ProbeHandler interface {
	GetAllProbes(ctx iris.Context)
	DeleteProbe(ctx iris.Context)
	CreateProbe(ctx iris.Context)
	UpdateProbe(ctx iris.Context)
	GetAllServices(ctx iris.Context)
}

const LogTag = "probeHandler"

type probeHandler struct {
	service service.ProbeService
	cfg     config.ZkOperatorConfig
}

func NewProbeHandler(service service.ProbeService) ProbeHandler {
	return &probeHandler{service: service}
}

func (p *probeHandler) GetAllProbes(ctx iris.Context) {
	ns := ctx.URLParam("ns")
	if common.IsEmpty(ns) {
		zklogger.Error(LogTag, "Namespace cannot be empty")
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Namespace cannot be empty"})
		return
	}

	resp, zkErr := p.service.GetAllProbes(ns)
	zkHttpResponse := zkhttp.ToZkResponse[response.CRDListResponse](iris.StatusOK, resp, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	err := ctx.JSON(zkHttpResponse)
	if err != nil {
		zklogger.Error(LogTag, "Error marshalling response", err)
		return
	}
}

func (p *probeHandler) DeleteProbe(ctx iris.Context) {
	ns := ctx.URLParam("ns")
	if common.IsEmpty(ns) {
		zklogger.Error(LogTag, "Namespace cannot be empty")
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Namespace cannot be empty"})
		return
	}

	zkErr := p.service.DeleteProbe(ns, ctx.Params().Get("name"))
	zkHttpResponse := zkhttp.ToZkResponse[any](iris.StatusOK, nil, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func (p *probeHandler) CreateProbe(ctx iris.Context) {
	ns := ctx.URLParam("ns")
	if common.IsEmpty(ns) {
		zklogger.Error(LogTag, "Namespace cannot be empty")
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Namespace cannot be empty"})
		return
	}

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

	zkErr := p.service.CreateProbe(ns, probeBody)
	zkHttpResponse := zkhttp.ToZkResponse[any](iris.StatusCreated, nil, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func (p *probeHandler) UpdateProbe(ctx iris.Context) {
	ns := ctx.URLParam("ns")
	if common.IsEmpty(ns) {
		zklogger.Error(LogTag, "Namespace cannot be empty")
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Namespace cannot be empty"})
		return
	}

	probeName := ctx.Params().Get("name")
	if common.IsEmpty(probeName) {
		zklogger.Error(LogTag, "Probe name cannot be empty")
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "Probe name cannot be empty"})
		return
	}

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

	zkErr := p.service.UpdateProbe(ns, probeName, probeBody)
	zkHttpResponse := zkhttp.ToZkResponse[any](iris.StatusOK, nil, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func (p *probeHandler) GetAllServices(ctx iris.Context) {
	resp, zkErr := p.service.GetAllServices()
	zkHttpResponse := zkhttp.ToZkResponse[response.ServiceListResponse](200, resp, nil, zkErr)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func readProbeRequest(ctx iris.Context) (request.UpsertProbeRequest, error) {
	var req request.UpsertProbeRequest
	body, err := ctx.GetBody()
	if err != nil {
		return req, err
	}

	err = yaml.Unmarshal(body, &req)
	if err != nil {
		return req, err
	}

	return req, nil
}

func validateProbeBody(probeBody request.UpsertProbeRequest) error {
	if common.IsEmpty(probeBody.Title) {
		zklogger.Error(LogTag, "Title cannot be empty")
		return errors.New("title cannot be empty")
	}

	if probeBody.Workloads == nil {
		zklogger.Error(LogTag, "Workloads cannot be empty")
		return errors.New("workloads cannot be empty")
	}

	if common.IsEmpty(probeBody.Filter.Type) {
		zklogger.Error(LogTag, "Filter cannot be empty")
		return errors.New("filter cannot be empty")
	}

	return nil
}
