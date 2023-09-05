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

	patches = append(patches, h.getContainerPatches(pod)...)

	logger.Debug(LOG_TAG, "The patches created are ", patches)

	return patches
}

func (h *Injector) modifyExistingCmd(overrideCmd []string) []string {
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

	logger.Debug(LOG_TAG, "otelSplitResultMap ", otelSplitResultMap)

	logger.Debug(LOG_TAG, "First time ", overrideCmd)

	prefix := "-Dotel."
	for i, cmd := range overrideCmd {
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
		overrideCmd[i] = strings.Join(cmdSplitArr, " ")
	}

	//Add missing flags
	missingFlagsArr := []string{}
	for name, value := range otelSplitResultMap {
		item := name + "=" + value
		missingFlagsArr = append(missingFlagsArr, item)
	}

	logger.Debug(LOG_TAG, "Second time ", overrideCmd)

	jarIndex := -1
	if len(missingFlagsArr) > 0 {
		for i, cmd := range overrideCmd {
			if strings.Contains(cmd, "-jar") {
				jarIndex = i
				break
			}
		}
	}

	logger.Debug(LOG_TAG, "Second time ", overrideCmd)

	overrideCmd = append(overrideCmd[:jarIndex], append(missingFlagsArr, overrideCmd[jarIndex:]...)...)

	return overrideCmd
}

// These patches orchestrate the container based on language.
func (h *Injector) getContainerPatches(pod *corev1.Pod) []map[string]interface{} {

	patches := make([]map[string]interface{}, 0)

	containers := pod.Spec.Containers

	shouldAddStandardPatches := false

	for index := range containers {

		container := &pod.Spec.Containers[index]

		language := h.ImageRuntimeCache.GetContainerLanguage(container)

		override := h.ImageRuntimeCache.GetOverrideForImage(container.Image)

		logger.Debug(LOG_TAG, "Found language ", language, " for container ", container.Name)

		shouldAddStandardPatchesToContainer := false

		switch language {
		case common.JavaProgrammingLanguage:
			//Adding env variable patch in case the prog language is java.
			javaEnvPatch := h.addJavaToolEnvPatch(container, index, override)
			patches = append(patches, javaEnvPatch...)
			orchLabelPatch := getZerokLabelPatch(common.ZkOrchOrchestrated)
			patches = append(patches, orchLabelPatch)
			shouldAddStandardPatchesToContainer = true
			shouldAddStandardPatches = true
		case common.NotYetProcessed:
			//Setting zk-status as in-process since there is not runtime info in redis.
			orchLabelPatch := getZerokLabelPatch(common.ZkOrchInProcess)
			patches = append(patches, orchLabelPatch)
		default:
			orchLabelPatch := getZerokLabelPatch(common.ZkOrchProcessed)
			patches = append(patches, orchLabelPatch)
		}

		if shouldAddStandardPatchesToContainer {
			addVolumeMount := h.getVolumeMount(index)

			patches = append(patches, addVolumeMount)

			if override != nil {
				patches = append(patches, h.addEnvOverridePatch(container, index, override.Env)...)

				cmdOverride := override.CmdOverride

				if len(cmdOverride) > 0 {
					newCmd := h.modifyExistingCmd(cmdOverride)
					if len(container.Command) > 0 {
						patch := map[string]interface{}{
							"op":    "replace",
							"path":  fmt.Sprintf("/spec/containers/%v/command", index),
							"value": newCmd,
						}
						patches = append(patches, patch)
					} else {
						patch := map[string]interface{}{
							"op":    "add",
							"path":  fmt.Sprintf("/spec/containers/%v/command", index),
							"value": newCmd,
						}
						patches = append(patches, patch)
					}
				}
			} else {
				logger.Debug(LOG_TAG, "Did not find any override for the image ", container.Image)
			}

			//Adding zk redis env patches
			patches = append(patches, h.getRedisEnvVarPatches(index)...)
		}

	}

	if shouldAddStandardPatches {
		standardPatches := h.getStandardPatches(pod)
		standardPatches = append(standardPatches, patches...)
		patches = standardPatches
	}

	return patches
}

func (h *Injector) getStandardPatches(pod *corev1.Pod) []map[string]interface{} {
	var patches []map[string]interface{}
	patches = append(patches, h.getInitContainerPatches(pod)...)
	patches = append(patches, h.getVolumePatch())
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

func (h *Injector) addJavaToolEnvPatch(container *corev1.Container, containerIndex int, override *v1alpha1.ImageOverride) []map[string]interface{} {
	envVars := container.Env
	envIndex := -1
	patches := []map[string]interface{}{}

	//If there are no env variables in container, adding an empty array first.
	if len(envVars) == 0 {
		patches = append(patches, h.addEnvObjectPatch(containerIndex))
	} else {
		envIndex = utils.GetIndexOfEnv(envVars, common.JavaToolOptions)
	}

	var patch map[string]interface{}

	// case 1: Override value present for java tool options.
	var overrideEnvVars []v1alpha1.EnvVar
	if override != nil {
		overrideEnvVars = override.Env
	}

	currentJavaToolOptions := ""
	if len(overrideEnvVars) > 0 {
		for _, overrideEnv := range overrideEnvVars {
			name := overrideEnv.Name
			value := overrideEnv.Value
			if name == common.JavaToolOptions {
				currentJavaToolOptions = value
				break
			}
		}
	} else {
		//Getting env vars from daemonset.
		runtime := h.ImageRuntimeCache.GetRuntimeForImage(container.Image)
		runtimeEnvVars := runtime.EnvMap
		runtimeJavaToolOpt, ok := runtimeEnvVars[common.JavaToolOptions]

		if envIndex > 0 {
			// case 2: No override present. But value present in pod spec.
			currentJavaToolOptions = container.Env[envIndex].Value
		} else if ok {
			// case 3: Command found in daemonset .
			currentJavaToolOptions = runtimeJavaToolOpt
		}
	}

	//Scenario where java_tool_options is not found in any of above scenarios.
	if len(currentJavaToolOptions) == 0 {
		patch = h.getAddEnvPatch(containerIndex, common.JavaToolOptions, h.Config.Instrumentation.OtelArgument)
	} else {
		splitResult := strings.Split(h.Config.Instrumentation.OtelArgument, " ")
		prefix := "-Dotel."
		for _, str := range splitResult {
			if strings.HasPrefix(str, prefix) {
				splitResultNew := strings.Split(str, "=")
				if len(splitResultNew) == 2 {
					flag := splitResultNew[0]
					value := splitResultNew[1]
					currentJavaToolOptions = h.ImageRuntimeCache.AddorUpdateFlags(currentJavaToolOptions, flag, value)
				}
			}
		}
		if envIndex == -1 {
			patch = h.getAddEnvPatch(containerIndex, common.JavaToolOptions, currentJavaToolOptions)
		} else {
			patch = h.getReplaceEnvPatch(containerIndex, envIndex, common.JavaToolOptions, currentJavaToolOptions)
		}
	}
	patches = append(patches, patch)
	return patches
}

func (h *Injector) addEnvOverridePatch(container *corev1.Container, containerIndex int, overrideEnv []v1alpha1.EnvVar) []map[string]interface{} {
	specEnvVars := container.Env
	envIndex := -1
	patches := []map[string]interface{}{}

	logger.Debug(LOG_TAG, "Override env vars ", overrideEnv)

	//If there are no env variables in container, adding an empty array first.
	if len(specEnvVars) == 0 && len(overrideEnv) > 0 {
		patches = append(patches, h.addEnvObjectPatch(containerIndex))
	}

	for _, overrideEnv := range overrideEnv {
		name := overrideEnv.Name
		value := overrideEnv.Value
		//Ignoring java_tool_versions to add a special handling for that.
		if name == common.JavaToolOptions {
			continue
		}
		var patch map[string]interface{}
		envIndex = utils.GetIndexOfEnv(specEnvVars, name)

		if envIndex == -1 {
			patch = h.getAddEnvPatch(containerIndex, name, value)
		} else {
			patch = h.getReplaceEnvPatch(containerIndex, envIndex, name, value)
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

func (h *Injector) getRedisEnvVarPatches(containerIndex int) []map[string]interface{} {
	var patches []map[string]interface{}
	patches = append(patches, h.getAddEnvPatch(containerIndex, "ZK_REDIS_HOSTNAME", h.Config.Redis.Host))
	patches = append(patches, h.getAddEnvPatch(containerIndex, "ZK_REDIS_PASSWORD", h.Config.Redis.Password))
	patches = append(patches, h.getAddEnvPatch(containerIndex, "ZK_REDIS_PORT", "6379"))
	patches = append(patches, h.getAddEnvPatch(containerIndex, "ZK_REDIS_DB", "3"))
	patches = append(patches, h.getAddEnvPatch(containerIndex, "ZK_REDIS_TTL", "900"))
	patches = append(patches, h.getAddEnvPatch(containerIndex, "ZK_REDIS_BATCH_SIZE", "100"))
	patches = append(patches, h.getAddEnvPatch(containerIndex, "ZK_REDIS_DURATION_MILLIS", "5000"))
	patches = append(patches, h.getAddEnvPatch(containerIndex, "ZK_REDIS_TIMER_SYNC_DURATION", "30000"))

	return patches
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
