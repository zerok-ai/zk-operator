package inject

import (
	"encoding/json"
	"fmt"
	common "github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/storage"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"strconv"
	"time"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Injector is a struct that implements an admission controller webhook for Kubernetes pods.
type Injector struct {
	ImageRuntimeHandler *storage.ImageRuntimeHandler
}

// GetEmptyResponse returns an empty admission response as a JSON byte array.
func (h *Injector) GetEmptyResponse(admissionReview v1.AdmissionReview) ([]byte, error) {
	ar := admissionReview.Request
	if ar != nil {
		admissionResponse := v1.AdmissionResponse{}
		admissionResponse.UID = ar.UID
		admissionResponse.Allowed = true
		patchType := v1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
		patches := make([]map[string]interface{}, 0)
		admissionResponse.Patch, _ = json.Marshal(patches)
		admissionResponse.Result = &metav1.Status{
			Status: "Success",
		}
		admissionReview.Response = &admissionResponse
		responseBody, err := json.Marshal(admissionReview)
		if err != nil {
			return nil, fmt.Errorf("error caught while marshalling response %v", err)
		}
		return responseBody, nil
	}
	return nil, fmt.Errorf("empty admission request")
}

// Inject takes a JSON byte array as input, which represents an admission review object, and returns an updated admission review object with patches applied to the pod.
func (h *Injector) Inject(body []byte) ([]byte, error) {
	admissionReviewObj := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &admissionReviewObj); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var pod *corev1.Pod

	responseBody := []byte{}
	admissionRequest := admissionReviewObj.Request
	admissionResponse := v1.AdmissionResponse{}
	emptyResponse, _ := h.GetEmptyResponse(admissionReviewObj)

	if admissionRequest != nil {

		if err := json.Unmarshal(admissionRequest.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}

		fmt.Printf("Got a request for POD = %s\n", pod.Name)

		admissionResponse.UID = admissionRequest.UID

		dt := time.Now()
		fmt.Println("Got request with uid ", admissionRequest.UID, " at time ", dt.String())
		admissionResponse.Allowed = true

		patchType := v1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType

		//Creating the patches to be applied on the pod.
		patches := h.getPatches(pod)

		var err error
		//Creating json patch to send in admission response.
		admissionResponse.Patch, err = json.Marshal(patches)

		if err != nil {
			fmt.Printf("Error caught while marshalling the patches %v.\n", err)
			//Sending empty response to let the pod creation happen without instrumentation.
			return emptyResponse, err
		}

		admissionResponse.Result = &metav1.Status{
			Status: "Success",
		}

		admissionReviewObj.Response = &admissionResponse

		responseBody, err = json.Marshal(admissionReviewObj)
		if err != nil {
			//Sending empty response to let the pod creation happen without instrumentation.
			return emptyResponse, err
		}
	}

	fmt.Printf("resp: %s\n", string(responseBody))

	return responseBody, nil
}

// This method returns all the patches to be applied on the pod.
func (h *Injector) getPatches(pod *corev1.Pod) []map[string]interface{} {
	patches := make([]map[string]interface{}, 0)

	//These set of patches will inject the init container.
	patches = append(patches, h.getInitContainerPatches(pod)...)

	//This patch for adding volume mount. This allows the main container access to otel agent.
	patches = append(patches, h.getVolumePatch())

	//These patchs orchestraces the container based on language.
	containerPatches := h.getContainerPatches(pod)
	patches = append(patches, containerPatches...)

	fmt.Printf("The patches created are %v.\n", patches)

	return patches
}

func (h *Injector) getContainerPatches(pod *corev1.Pod) []map[string]interface{} {

	patches := make([]map[string]interface{}, 0)

	containers := pod.Spec.Containers

	for index := range containers {

		container := &pod.Spec.Containers[index]

		language := h.ImageRuntimeHandler.GetContainerLanguage(container, pod)

		fmt.Printf("Found language %v for container %v\n", language, container.Name)

		switch language {
		case common.JavaProgrammingLanguage:
			//Adding env variable patch in case the prog language is java.
			javaToolsPatch := modifyJavaToolsEnvVariablePatch(container, index)
			patches = append(patches, javaToolsPatch...)
			orchLabelPatch := getZerokLabelPatch(common.ZkOrchOrchestrated)
			patches = append(patches, orchLabelPatch)
		case common.NotYetProcessed:
			//Setting zk-status as in-process since there is not runtime info in redis.
			orchLabelPatch := getZerokLabelPatch(common.ZkOrchInProcess)
			patches = append(patches, orchLabelPatch)
		default:
			orchLabelPatch := getZerokLabelPatch(common.ZkOrchProcessed)
			patches = append(patches, orchLabelPatch)
		}

		addVolumeMount := h.getVolumeMount(index)

		patches = append(patches, addVolumeMount)

	}

	return patches
}

func modifyJavaToolsEnvVariablePatch(container *corev1.Container, containerIndex int) []map[string]interface{} {
	envVars := container.Env
	envIndex := -1
	patches := []map[string]interface{}{}

	//If there are no env variables in container, adding an empty array first.
	if len(envVars) == 0 {
		envInitialize := map[string]interface{}{
			"op":    "add",
			"path":  fmt.Sprintf("/spec/containers/%v/env", containerIndex),
			"value": []corev1.EnvVar{},
		}
		patches = append(patches, envInitialize)
	} else {
		envIndex = utils.GetIndexOfEnv(envVars, common.JavalToolOptions)
	}

	var patch map[string]interface{}
	//Scenario where java_tool_options is not present.
	if envIndex == -1 {
		patch = map[string]interface{}{
			"op":   "add",
			"path": fmt.Sprintf("/spec/containers/%v/env/-", containerIndex),
			"value": corev1.EnvVar{
				Name:  common.JavalToolOptions,
				Value: common.OtelArgument,
			},
		}

	} else {
		//Scenario where java_tool_options is already present.
		patch = map[string]interface{}{
			"op":   "replace",
			"path": fmt.Sprintf("/spec/containers/%v/env/%v", containerIndex, envIndex),
			"value": corev1.EnvVar{
				Name:  common.JavalToolOptions,
				Value: container.Env[envIndex].Value + common.OtelArgument,
			},
		}
	}
	patches = append(patches, patch)
	return patches
}

func getZerokLabelPatch(value string) map[string]interface{} {
	labelPod := map[string]interface{}{
		"op":    "replace",
		"path":  common.ZkOrchPath,
		"value": value,
	}
	return labelPod
}

func (*Injector) getVolumeMount(i int) map[string]interface{} {
	addVolumeMount := map[string]interface{}{
		"op":   "add",
		"path": "/spec/containers/" + strconv.Itoa(i) + "/volumeMounts/-",
		"value": corev1.VolumeMount{
			MountPath: "/opt/zerok",
			Name:      "zerok-init",
		},
	}
	return addVolumeMount
}

func (h *Injector) getVolumePatch() map[string]interface{} {
	addVolume := map[string]interface{}{
		"op":   "add",
		"path": "/spec/volumes/-",
		"value": corev1.Volume{
			Name: "zerok-init",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	return addVolume
}

func (h *Injector) getInitContainerPatches(pod *corev1.Pod) []map[string]interface{} {
	p := make([]map[string]interface{}, 0)

	if pod.Spec.InitContainers == nil {
		initInitialize := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/initContainers",
			"value": []corev1.Container{},
		}

		p = append(p, initInitialize)
	}

	addInitContainer := map[string]interface{}{
		"op":   "add",
		"path": "/spec/initContainers/-",
		"value": &corev1.Container{
			Name:            "zerok-init",
			Command:         []string{"cp", "-r", "/opt/zerok/.", "/opt/temp"},
			Image:           "us-west1-docker.pkg.dev/zerok-dev/stage/init-container:test",
			ImagePullPolicy: corev1.PullAlways,
			VolumeMounts: []corev1.VolumeMount{
				{
					MountPath: "/opt/temp",
					Name:      "zerok-init",
				},
			},
		},
	}

	p = append(p, addInitContainer)

	return p
}
