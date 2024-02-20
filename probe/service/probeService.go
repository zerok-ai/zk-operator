package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"github.com/zerok-ai/zk-operator/probe/model/request"
	"github.com/zerok-ai/zk-operator/probe/model/response"
	"github.com/zerok-ai/zk-operator/store"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/zkerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
)

var LogTag = "probeService"

type ProbeService interface {
	GetAllProbes() (response.CRDListResponse, *zkerrors.ZkError)
	DeleteProbe(name string) *zkerrors.ZkError
	UpdateProbe(name string, probe request.UpsertProbeRequest) *zkerrors.ZkError
	CreateProbe(request.UpsertProbeRequest) *zkerrors.ZkError
	GetAllServices() (response.ServiceListResponse, *zkerrors.ZkError)
}

type probeService struct {
	serviceStore store.ServiceStore
}

func NewProbeService(serviceStore *store.ServiceStore) ProbeService {
	return &probeService{
		serviceStore: *serviceStore,
	}
}

func (p *probeService) GetAllProbes() (response.CRDListResponse, *zkerrors.ZkError) {
	var probeList response.CRDListResponse
	clientSet, err := createDynamicClient()
	if err != nil {
		zklogger.Error("Error while creating k8s client config", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return probeList, &zkErr
	}

	crdList, err := clientSet.Resource(utils.SchemaGroupVersionKindForResource()).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		zklogger.Error("Error while listing CRDs", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return probeList, &zkErr
	}

	crds := make([]response.ProbeResponse, 0)
	for _, item := range crdList.Items {
		spec, _, _ := unstructured.NestedMap(item.Object, "spec")
		var myCRD response.ProbeResponse
		jsonStr, err := json.Marshal(spec)
		if err != nil {
			zklogger.Error("Error while marshalling CRD struct to YAML", err)
			zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
			return probeList, &zkErr
		}

		err = json.Unmarshal(jsonStr, &myCRD)
		if err != nil {
			zklogger.Error("Error while unmarshalling YAML to CRD struct", err)
			zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
			return probeList, &zkErr
		}

		crds = append(crds, myCRD)
	}

	probeList.CRDList = crds
	return probeList, nil
}

func (p *probeService) DeleteProbe(name string) *zkerrors.ZkError {
	// Create a dynamic client
	clientSet, err := createDynamicClient()
	if err != nil {
		zklogger.Error("Error while creating k8s client config", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return &zkErr
	}

	zklogger.Error(LogTag, "Name: ", name)

	crd, err := getCRD(clientSet, name)
	if err != nil {
		zklogger.Error(LogTag, "Error while getting CRD", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	zklogger.Error(LogTag, "CRD name: ", crd.GetName())

	if crd.GetName() != name {
		zklogger.Error(LogTag, "Error while deleting CRD. Name in the CRD and probe struct do not match", nil)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return &zkErr
	}

	err = deleteCRD(clientSet, name)
	if err != nil {
		zklogger.Error("Error while deleting CRD", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	zklogger.Info(LogTag, "CRD deleted successfully\n")
	return nil
}

func (p *probeService) CreateProbe(probe request.UpsertProbeRequest) *zkerrors.ZkError {
	unstructuredObj, err := convertToUnstructured(probe)
	if err != nil {
		zklogger.Error("Error while converting to unstructured object", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	// Create a dynamic client
	clientSet, err := createDynamicClient()
	if err != nil {
		zklogger.Error("Error while creating k8s client config", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return &zkErr
	}

	_, err = clientSet.Resource(utils.SchemaGroupVersionKindForResource()).Namespace("zk-client").Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		zklogger.Error("Error while creating CRD", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	if err != nil {
		zklogger.Error("Error while creating CRD", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	zklogger.Info(LogTag, "CRD created successfully\n")
	return nil
}

func (p *probeService) UpdateProbe(probeName string, probe request.UpsertProbeRequest) *zkerrors.ZkError {
	clientSet, err := createDynamicClient()
	if err != nil {
		zklogger.Error("Error while creating k8s client config", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return &zkErr
	}

	zklogger.Error(LogTag, "Name: ", probeName)

	crd, err := getCRD(clientSet, probeName)
	if err != nil {
		zklogger.Error(LogTag, "Error while getting CRD", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	zklogger.Error(LogTag, "CRD name: ", crd.GetName())

	if crd.GetName() != probeName {
		zklogger.Error(LogTag, "Error while deleting CRD. Name in the CRD and probe struct do not match", nil)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return &zkErr
	}

	unstructuredObj, err := convertToUnstructured(probe)
	unstructuredObj.SetResourceVersion(crd.GetResourceVersion())

	if err != nil {
		zklogger.Error("Error while converting to unstructured object", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	_, err = clientSet.Resource(utils.SchemaGroupVersionKindForResource()).Namespace("zk-client").Update(context.TODO(), &unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
	}

	if err != nil {
		zklogger.Error("Error while creating CRD", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	zklogger.Info(LogTag, "CRD updated successfully\n")
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

func createDynamicClient() (*dynamic.DynamicClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if home := homedir.HomeDir(); home != "" {
			config, err = rest.InClusterConfig()
			if err != nil {
				zklogger.Error("Error while creating k8s client config", err)
				return nil, fmt.Errorf("failed to create Kubernetes client config: %v", err)
			}
		}
	}

	clientSet, err := dynamic.NewForConfig(config)
	if err != nil {
		zklogger.Error("Error while creating k8s client config", err)
		return nil, fmt.Errorf("failed to create Kubernetes client config: %v", err)
	}

	return clientSet, nil
}

func getCRD(clientSet *dynamic.DynamicClient, name string) (*unstructured.Unstructured, error) {
	crd, err := clientSet.Resource(utils.SchemaGroupVersionKindForResource()).Namespace("zk-client").Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		zklogger.Error("Error while getting CRD", err)
		return nil, fmt.Errorf("failed to get CRD: %v", err)
	}

	return crd, nil
}

func deleteCRD(clientSet *dynamic.DynamicClient, name string) error {
	err := clientSet.Resource(utils.SchemaGroupVersionKindForResource()).Namespace("zk-client").Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		zklogger.Error("Error while deleting CRD", err)
		return fmt.Errorf("failed to delete CRD: %v", err)
	}

	return nil
}

func convertToUnstructured(probeRequest request.UpsertProbeRequest) (unstructured.Unstructured, error) {
	var zkProbeSpec v1alpha1.ZerokProbeSpec
	zkProbeSpec.Title = probeRequest.Title
	zkProbeSpec.Enabled = probeRequest.Enabled

	workloadByteArr, err := json.Marshal(probeRequest.Workloads)
	if err != nil {
		zklogger.Error("Error while marshalling workloads", err)
		return unstructured.Unstructured{}, err
	}

	err = json.Unmarshal(workloadByteArr, &zkProbeSpec.Workloads)
	if err != nil {
		zklogger.Error("Error while unmarshalling workloads", err)
		return unstructured.Unstructured{}, err
	}

	filterByteArr, err := json.Marshal(probeRequest.Filter)
	if err != nil {
		zklogger.Error("Error while marshalling filter", err)
		return unstructured.Unstructured{}, err
	}

	err = json.Unmarshal(filterByteArr, &zkProbeSpec.Filter)
	if err != nil {
		zklogger.Error("Error while unmarshalling filter", err)
		return unstructured.Unstructured{}, err
	}

	var unstructuredObj unstructured.Unstructured
	probeReqStr, err := json.Marshal(zkProbeSpec)
	if err != nil {
		zklogger.Error("Error while marshalling probe request", err)
		return unstructuredObj, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(probeReqStr, &data)
	unstructuredObj = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", utils.Group, utils.Version),
			"kind":       utils.ZeroKProbeKind,
			"metadata": map[string]interface{}{
				"name": probeRequest.Title,
			},
			"spec": data,
		},
	}
	unstructuredObj.SetName(probeRequest.Title)

	unstructuredObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   utils.Group,
		Version: utils.Version,
		Kind:    utils.ZeroKProbeKind,
	})

	return unstructuredObj, nil
}
