package inject

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/api/v1alpha1"
	common "github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/storage"
	"github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var LOG_TAG = "inject"

// Injector is a struct that implements an admission controller webhook for Kubernetes pods.
type Injector struct {
	ImageRuntimeCache *storage.ImageRuntimeCache
	Config            config.ZkOperatorConfig
	InitContainerData *config.AppInitContainerData
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

	admissionRequest := admissionReviewObj.Request
	admissionResponse := v1.AdmissionResponse{}
	emptyResponse, _ := h.GetEmptyResponse(admissionReviewObj)

	if admissionRequest == nil {
		return emptyResponse, fmt.Errorf("admission request is nil")
	}

	if err := json.Unmarshal(admissionRequest.Object.Raw, &pod); err != nil {
		return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
	}

	logger.Debug(LOG_TAG, "Got a request for POD ", pod.Name)

	admissionResponse.UID = admissionRequest.UID

	dt := time.Now()
	logger.Debug(LOG_TAG, "Got request with uid ", admissionRequest.UID, " at time ", dt.String())
	admissionResponse.Allowed = true

	patchType := v1.PatchTypeJSONPatch
	admissionResponse.PatchType = &patchType

	//Creating the patches to be applied on the pod.
	patches := h.getPatches(pod)

	var err error
	//Creating json patch to send in admission response.
	admissionResponse.Patch, err = json.Marshal(patches)

	if err != nil {
		logger.Debug(LOG_TAG, "Error caught while marshalling the patches ", err)
		//Sending empty response to let the pod creation happen without instrumentation.
		return emptyResponse, err
	}

	admissionResponse.Result = &metav1.Status{
		Status: "Success",
	}

	admissionReviewObj.Response = &admissionResponse

	responseBody, err := json.Marshal(admissionReviewObj)
	if err != nil {
		//Sending empty response to let the pod creation happen without instrumentation.
		return emptyResponse, err
	}

	return responseBody, nil
}

// This method returns all the patches to be applied on the pod.
func (h *Injector) getPatches(pod *corev1.Pod) []map[string]interface{} {

	patches := make([]map[string]interface{}, 0)

	patches = append(patches, h.getInitContainerPatches(pod)...)
	patches = append(patches, h.getVolumePatch())
	patches = append(patches, h.getContainerPatches(pod)...)

	logger.Debug(LOG_TAG, "The patches created are ", patches)

	return patches
}

func (h *Injector) modifyExistingCmd(existingCmd []string) []string {
	var newCmd []string
	otelSplitResult := strings.Split(h.Config.Instrumentation.OtelArgument, " ")
	otelSplitResultMap := make(map[string]string)
	for _, result := range otelSplitResult {
		otelSplitResultItem := strings.Split(result, "=")
		if len(otelSplitResultItem) == 2 {
			flag := otelSplitResultItem[0]
			value := otelSplitResultItem[1]
			otelSplitResultMap[flag] = value
		}
	}

	prefix := "-Dotel."
	for i, cmd := range existingCmd {
		cmdSplitArr := strings.Split(cmd, " ")
		for i, cmdSplit := range cmdSplitArr {
			if strings.HasPrefix(cmdSplit, prefix) {
				cmdSplitItems := strings.Split(cmdSplit, "=")
				if len(cmdSplitItems) == 2 {
					existingFlag := cmdSplitItems[0]
					existingValue := cmdSplitItems[1]
					otelValue, ok := otelSplitResultMap[existingFlag]
					if ok {
						existingValue = existingValue + "," + otelValue
						delete(otelSplitResultMap, existingFlag)
					}
					newCmdItem := fmt.Sprintf("%s=%s", existingFlag, existingValue)
					cmdSplitArr[i] = newCmdItem
				}
			}
		}
		existingCmd[i] = strings.Join(cmdSplitArr, "")
	}

	return newCmd
}

// These patches orchestrate the container based on language.
func (h *Injector) getContainerPatches(pod *corev1.Pod) []map[string]interface{} {

	patches := make([]map[string]interface{}, 0)

	containers := pod.Spec.Containers

	for index := range containers {

		container := &pod.Spec.Containers[index]

		language := h.ImageRuntimeCache.GetContainerLanguage(container)

		override := h.ImageRuntimeCache.GetOverrideForImage(container.Image)

		logger.Debug(LOG_TAG, "Found language ", language, " for container ", container.Name)

		switch language {
		case common.JavaProgrammingLanguage:
			//Adding env variable patch in case the prog language is java.
			javaEnvPatch := h.addJavaToolEnvPatch(container, index)
			patches = append(patches, javaEnvPatch...)
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

		//TODO: Check if we are also considering the env values that we got from daemonset.
		patches = append(patches, h.addEnvOverridePatch(container, index, override.Env)...)

		cmdOverride := override.CmdOverride
		runtime := h.ImageRuntimeCache.GetRuntimeForImage(container.Image)

		//TODO: What is the role of zk-override and user-override here?
		//TODO: Should we explicitly ask users to add zk-override for all of them?
		if len(cmdOverride) > 0 {
			//We will have to make changes to cmdOverride and add it here.
			//TODO: Which one should take precedence here?
			if len(container.Command) > 0 {
				//Container cmd already present
				//We will have to give a replace patch here.
				//Create a new command and create a replace patch.
				newCmd := h.modifyExistingCmd(container.Command)
				patch := map[string]interface{}{
					"op":    "replace",
					"path":  fmt.Sprintf("/spec/containers/%v/command", index),
					"value": newCmd,
				}
				patches = append(patches, patch)
			} else {
				//Container cmd not present.
				//Take the command we got from runtime and change that.
				//We will have to give an add patch here.
				newCmd := h.modifyExistingCmd(runtime.Cmd)
				patch := map[string]interface{}{
					"op":    "add",
					"path":  fmt.Sprintf("/spec/containers/%v/command", index),
					"value": newCmd,
				}
				patches = append(patches, patch)
			}
		}
	}

	return patches
}

func (h *Injector) getAddEnvPatch(containerIndex int, name, value string) map[string]interface{} {
	patch := map[string]interface{}{
		"op":   "add",
		"path": fmt.Sprintf("/spec/containers/%v/env/-", containerIndex),
		"value": corev1.EnvVar{
			Name:  name,
			Value: value,
		},
	}
	return patch
}

func (h *Injector) getReplaceEnvPatch(containerIndex, envIndex int, name, value string) map[string]interface{} {
	patch := map[string]interface{}{
		"op":   "replace",
		"path": fmt.Sprintf("/spec/containers/%v/env/%v", containerIndex, envIndex),
		"value": corev1.EnvVar{
			Name:  name,
			Value: value,
		},
	}
	return patch
}

func (h *Injector) addEnvObjectPatch(containerIndex int) map[string]interface{} {
	envInitialize := map[string]interface{}{
		"op":    "add",
		"path":  fmt.Sprintf("/spec/containers/%v/env", containerIndex),
		"value": []corev1.EnvVar{},
	}
	return envInitialize
}

func (h *Injector) addJavaToolEnvPatch(container *corev1.Container, containerIndex int) []map[string]interface{} {
	envVars := container.Env
	envIndex := -1
	patches := []map[string]interface{}{}

	//If there are no env variables in container, adding an empty array first.
	if len(envVars) == 0 {
		patches = append(patches, h.addEnvObjectPatch(containerIndex))
	} else {
		envIndex = utils.GetIndexOfEnv(envVars, common.JavalToolOptions)
	}

	var patch map[string]interface{}
	//Scenario where java_tool_options is not present.
	if envIndex == -1 {
		patch = h.getAddEnvPatch(containerIndex, common.JavalToolOptions, h.Config.Instrumentation.OtelArgument)
	} else {
		//Scenario where java_tool_options is already present.
		splitResult := strings.Split(h.Config.Instrumentation.OtelArgument, " ")
		prefix := "-Dotel."
		existingValue := container.Env[envIndex].Value
		//TODO: Check if we are also handling the case where the flag is not present?
		for _, str := range splitResult {
			if strings.HasPrefix(str, prefix) {
				splitResultNew := strings.Split(str, "=")
				if len(splitResultNew) == 2 {
					flag := splitResultNew[0]
					value := splitResultNew[1]
					existingValue = h.ImageRuntimeCache.AddorUpdateFlags(existingValue, flag, value)
				}
			}
		}
		patch = h.getReplaceEnvPatch(containerIndex, envIndex, common.JavalToolOptions, existingValue)
	}
	patches = append(patches, patch)
	return patches
}

func (h *Injector) addEnvOverridePatch(container *corev1.Container, containerIndex int, overrideEnv []v1alpha1.EnvVar) []map[string]interface{} {
	//TODO: Do we need to add check for user-override value here?
	envVars := container.Env
	envIndex := -1
	patches := []map[string]interface{}{}
	isAdd := false

	//If there are no env variables in container, adding an empty array first.
	if len(envVars) == 0 {
		patches = append(patches, h.addEnvObjectPatch(containerIndex))
		isAdd = true
	}

	for _, overrideEnv := range overrideEnv {
		name := overrideEnv.Name
		value := overrideEnv.Value
		var patch map[string]interface{}
		if isAdd {
			patch = h.getAddEnvPatch(containerIndex, name, value)
		} else {
			envIndex = utils.GetIndexOfEnv(envVars, name)
			if envIndex == -1 {
				patch = h.getAddEnvPatch(containerIndex, name, value)
			} else {
				patch = h.getReplaceEnvPatch(containerIndex, envIndex, name, value)
			}
		}
		patches = append(patches, patch)
	}
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

// This patch for adding volume mount. This allows the main container access to otel agent.
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

// This patch for adding volume mount. This allows the main container access to otel agent.
func (h *Injector) getInitContainerPatches(pod *corev1.Pod) []map[string]interface{} {
	p := make([]map[string]interface{}, 0)

	initImage := h.InitContainerData.Image
	initTag := h.InitContainerData.Tag
	if len(initImage) == 0 || len(initTag) == 0 {
		return p
	}

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
			Image:           initImage + ":" + initTag,
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
