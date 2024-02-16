package service

import (
	"context"
	"encoding/json"
	"fmt"
	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"github.com/zerok-ai/zk-operator/probe/model/response"
	"github.com/zerok-ai/zk-operator/store"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	"github.com/zerok-ai/zk-utils-go/zkerrors"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
)

var LogTag = "probeService"

type ProbeService interface {
	GetAllProbes() (response.CRDListResponse, *zkerrors.ZkError)
	DeleteProbe(name string) *zkerrors.ZkError
	UpdateProbe(probe operatorv1alpha1.ZerokProbeSpec) *zkerrors.ZkError
	CreateProbe(operatorv1alpha1.ZerokProbeSpec) *zkerrors.ZkError
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

	fmt.Printf("CRD list: %v\n", crdList)

	crds := make([]model.Scenario, 0)
	for _, item := range crdList.Items {
		spec, _, _ := unstructured.NestedMap(item.Object, "spec")
		fmt.Println("-----------------" + item.GetName())
		var myCRD model.Scenario
		jsonStr, err := yaml.Marshal(spec)
		fmt.Println(jsonStr)
		if err != nil {
			zklogger.Error("Error while marshalling CRD struct to YAML", err)
			zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
			return probeList, &zkErr
		}

		err = yaml.Unmarshal(jsonStr, &myCRD)
		if err != nil {
			zklogger.Error("Error while unmarshalling YAML to CRD struct", err)
			zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
			return probeList, &zkErr
		}

		crds = append(crds, myCRD)
	}

	a, _ := json.Marshal(crds)
	fmt.Println(string(a))

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

	zklogger.Error(LogTag, "CRD name222222: ", crd.GetName())

	err = deleteCRD(clientSet, name)
	if err != nil {
		zklogger.Error("Error while deleting CRD", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	fmt.Printf("CRD deleted successfully\n")
	return nil
}

func (p *probeService) UpdateProbe(probe operatorv1alpha1.ZerokProbeSpec) *zkerrors.ZkError {
	yamlData, err := yaml.Marshal(probe)
	if err != nil {
		zklogger.Error("Error while marshalling CRD struct to YAML", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	var unstructuredObj unstructured.Unstructured
	err = yaml.Unmarshal(yamlData, &unstructuredObj)
	if err != nil {
		zklogger.Error("Error while unmarshalling YAML to unstructured object", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	if unstructuredObj.GetName() != probe.Title {
		zklogger.Error("Error while updating CRD. Name in the CRD and probe struct do not match", nil)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return &zkErr
	}

	// Create a dynamic client
	clientSet, err := createDynamicClient()
	if err != nil {
		zklogger.Error("Error while creating k8s client config", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return &zkErr
	}

	_, err = upsertCRD(clientSet, &unstructuredObj)
	if err != nil {
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	fmt.Printf("CRD updated successfully\n")
	return nil
}

func (p *probeService) CreateProbe(probe operatorv1alpha1.ZerokProbeSpec) *zkerrors.ZkError {
	yamlData, err := yaml.Marshal(probe)
	if err != nil {
		zklogger.Error("Error while marshalling CRD struct to YAML", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	var unstructuredObj unstructured.Unstructured
	err = yaml.Unmarshal(yamlData, &unstructuredObj)
	if err != nil {
		zklogger.Error("Error while unmarshalling YAML to unstructured object", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	unstructuredObj.SetName(probe.Title)

	// Create a dynamic client
	clientSet, err := createDynamicClient()
	if err != nil {
		zklogger.Error("Error while creating k8s client config", err)
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, nil)
		return &zkErr
	}

	_, err = upsertCRD(clientSet, &unstructuredObj)
	if err != nil {
		zkErr := zkerrors.ZkErrorBuilder{}.Build(zkerrors.ZkErrorInternalServer, err.Error())
		return &zkErr
	}

	fmt.Printf("CRD created successfully\n")
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

func upsertCRD(clientSet *dynamic.DynamicClient, crdUpsertReq *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	// Create the CRD
	crd, err := clientSet.Resource(utils.SchemaGroupVersionKindForResource()).Namespace("zk-client").Create(context.Background(), crdUpsertReq, metav1.CreateOptions{})
	if err != nil {
		zklogger.Error("Error while creating CRD", err)
		return nil, fmt.Errorf("failed to create CRD: %v", err)
	}

	return crd, nil
}
