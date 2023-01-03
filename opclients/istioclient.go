package opclients

import (
	"context"
	"fmt"
	"log"

	"k8s.io/client-go/rest"

	v1alpha3Spec "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func ApplyEnvoyConfig() {
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
			Name:      "gateway-access-log-new",
			Namespace: "default",
		},
		Spec: v1alpha3Spec.EnvoyFilter{
			WorkloadSelector: &v1alpha3Spec.WorkloadSelector{
				Labels: map[string]string{
					"zk-status":     "enabled",
					"zk-route-mark": "soak",
				},
			},
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
						Value:     GetRateLimiterValueStruct(),
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
