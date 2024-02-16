package service

import (
	"github.com/kataras/iris/v12"
	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/probe/model/response"
	"github.com/zerok-ai/zk-operator/store"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/zkerrors"
)

type ProbeService interface {
	GetAllProbes(ctx iris.Context)
	DeleteProbe(ctx iris.Context)
	UpdateProbe(ctx iris.Context)
	CreateProbe(operatorv1alpha1.ZerokProbeSpec) error
	GetAllServices() (response.ServiceListResponse, *zkerrors.ZkError)
}

type probeService struct {
	serviceStore store.ServiceStore
}

func (p *probeService) GetAllProbes(ctx iris.Context) {
	//TODO implement me
	panic("implement me")
}

func (p *probeService) DeleteProbe(ctx iris.Context) {
	//TODO implement me
	panic("implement me")
}

func (p *probeService) UpdateProbe(ctx iris.Context) {
	//TODO implement me
	panic("implement me")
}

func (p *probeService) CreateProbe(operatorv1alpha1.ZerokProbeSpec) error {
	//zkProbeSpecs := v1alpha1.ZerokProbeSpec{
	//	Title:     scenario.Title,
	//	Enabled:   true,
	//	Workloads: scenario.Workloads,
	//}
	//zkProbe := v1alpha1.ZerokProbe{
	//	TypeMeta:   metav1.TypeMeta{},
	//	ObjectMeta: metav1.ObjectMeta{},
	//	Spec:       v1alpha1.ZerokProbeSpec{},
	//	Status:     v1alpha1.ZerokProbeStatus{},
	//}

	//fmt.Print(scenario.Id)
	return nil
}

func (p *probeService) GetAllServices() (response.ServiceListResponse, *zkerrors.ZkError) {
	var serviceListResponse response.ServiceListResponse
	services, err := p.serviceStore.GetServices()
	if err != nil {
		zklogger.Error("Error while getting services from store", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return serviceListResponse, &zkErr
	}

	serviceListResponse.Services = services
	return serviceListResponse, nil
}

func NewProbeService(serviceStore *store.ServiceStore) ProbeService {
	return &probeService{
		serviceStore: *serviceStore,
	}
}
