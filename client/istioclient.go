package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1alpha3Spec "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func GetLogValueStruct() *structpb.Struct {
	accessFile := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"@type": {
				Kind: &structpb.Value_StringValue{
					StringValue: "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog",
				},
			},
			"path": {
				Kind: &structpb.Value_StringValue{
					StringValue: "/dev/stdout",
				},
			},
			"format": {
				Kind: &structpb.Value_StringValue{
					StringValue: "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \n",
				},
			},
		},
	}
	acessLog := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": {
				Kind: &structpb.Value_StringValue{
					StringValue: "envoy.access_loggers.file",
				},
			},
			"typed_config": {
				Kind: &structpb.Value_StructValue{
					StructValue: accessFile,
				},
			},
		},
	}
	acessLogValue := &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: acessLog,
		},
	}
	typedConfig := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"@type": {
				Kind: &structpb.Value_StringValue{
					StringValue: "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
				},
			},
			"access_log": {
				Kind: &structpb.Value_ListValue{
					ListValue: &structpb.ListValue{
						Values: []*structpb.Value{acessLogValue},
					},
				},
			},
		},
	}
	valueStruct := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"typed_config": {
				Kind: &structpb.Value_StructValue{
					StructValue: typedConfig,
				},
			},
		},
	}
	return valueStruct
}

func GetRateLimiterStruct() *structpb.Struct {
	valueStruct := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": {
				Kind: &structpb.Value_StringValue{
					StringValue: "envoy.filters.http.local_ratelimit",
				},
			},
		},
	}
	return valueStruct
}

func ApplyEnvoyConfig() {
	PrintPodsInCluster()
	fmt.Println("Starting")
	ic := GetIstioClient()
	fmt.Println("Create Istio client.")
	envoyFilterCrd := GetEnvoyFilterCrd()
	fmt.Println("Create envoy filter crd")
	fmt.Println("EnvoyFilterCrd is " + envoyFilterCrd.Spec.String())
	_, err := ic.NetworkingV1alpha3().EnvoyFilters(envoyFilterCrd.Namespace).Create(context.Background(), envoyFilterCrd, metav1.CreateOptions{})
	if err == nil {
		fmt.Println("Envoy Filter applied successfully.")
	} else {
		fmt.Println(err)
	}
	fmt.Println("Applied envoy filter crd.")
}

func GetEnvoyFilterCrd() *v1alpha3.EnvoyFilter {
	envoyFilterCrd := &v1alpha3.EnvoyFilter{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "EnvoyFilter",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gateway-access-log",
			Namespace: "default",
		},
		Spec: v1alpha3Spec.EnvoyFilter{
			ConfigPatches: []*v1alpha3Spec.EnvoyFilter_EnvoyConfigObjectPatch{
				{
					ApplyTo: v1alpha3Spec.EnvoyFilter_NETWORK_FILTER,
					Match: &v1alpha3Spec.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: v1alpha3Spec.EnvoyFilter_ANY,
						ObjectTypes: &v1alpha3Spec.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &v1alpha3Spec.EnvoyFilter_ListenerMatch{
								FilterChain: &v1alpha3Spec.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Filter: &v1alpha3Spec.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.http_connection_manager",
									},
								},
							},
						},
					},
					Patch: &v1alpha3Spec.EnvoyFilter_Patch{
						Operation: v1alpha3Spec.EnvoyFilter_Patch_MERGE,
						Value:     GetLogValueStruct(),
					},
				},
				{
					ApplyTo: v1alpha3Spec.EnvoyFilter_HTTP_FILTER,
					Match: &v1alpha3Spec.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: v1alpha3Spec.EnvoyFilter_SIDECAR_INBOUND,
						ObjectTypes: &v1alpha3Spec.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &v1alpha3Spec.EnvoyFilter_ListenerMatch{
								FilterChain: &v1alpha3Spec.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Filter: &v1alpha3Spec.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.http_connection_manager",
									},
								},
							},
						},
					},
					Patch: &v1alpha3Spec.EnvoyFilter_Patch{
						Operation: v1alpha3Spec.EnvoyFilter_Patch_INSERT_BEFORE,
						Value:     GetRateLimiterStruct(),
					},
				},
			},
		},
	}
	return envoyFilterCrd
}

func GetIstioClient() *versionedclient.Clientset {

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to create k8s rest client: %s", err)
	}

	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}
	return ic
}

func PrintPodsInCluster() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	k8sClient := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	name := "default"

	services, err := k8sClient.Services(name).List(context.Background(), listOptions)

	for _, service := range services.Items {
		if name == "default" && service.GetName() == "kubernetes" {
			continue
		}
		fmt.Println("namespace", name, "serviceName:", service.GetName(), "serviceKind:", service.Kind, "serviceLabels:", service.GetLabels(), service.Spec.Ports, "serviceSelector:", service.Spec.Selector)

		// labels.Parser
		set := labels.Set(service.Spec.Selector)

		if pods, err := k8sClient.Pods(name).List(context.Background(), metav1.ListOptions{LabelSelector: set.AsSelector().String()}); err != nil {
			fmt.Printf("List Pods of service[%s] error:%v\n", service.GetName(), err)
		} else {
			for _, pod := range pods.Items {
				fmt.Println("Pod", pod.GetName(), pod.Spec.NodeName, pod.Spec.Containers)
				payload := []patchStringValue{{
					Op:    "replace",
					Path:  "/metadata/labels/testLabel",
					Value: "897889",
				}}
				payloadBytes, _ := json.Marshal(payload)

				_, updateErr := k8sClient.Pods(pod.GetNamespace()).Patch(context.Background(), pod.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
				if updateErr == nil {
					fmt.Println(fmt.Sprintf("Pod %s labelled successfully.", pod.GetName()))
				} else {
					fmt.Println(updateErr)
				}
			}
		}
	}

}
