package storage

import (
	"fmt"
	"github.com/zerok-ai/zk-operator/api/v1alpha1"
	common "github.com/zerok-ai/zk-operator/internal/common"
	utils "github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	"sync"
	"time"

	"github.com/zerok-ai/zk-operator/internal/config"
	corev1 "k8s.io/api/core/v1"
)

var LOG_TAG = "ImageRuntimeCache"

type ImageRuntimeCache struct {
	//Key will be regex and value will be *v1alpha1.ImageOverride
	ImageOverrideMap  *sync.Map
	ImageRuntimeMap   *sync.Map
	RuntimeMapVersion int64
	ImageStore        *ImageStore
	ticker            *zktick.TickerTask
}

func (h *ImageRuntimeCache) StartPeriodicSync() {
	//Sync first time on pod start
	err := h.SyncDataFromRedis()
	if err != nil {
		logger.Error(LOG_TAG, "Error while syncing data from redis ", err)
	}
	h.ticker.Start()
}

func (h *ImageRuntimeCache) periodicSync() {
	logger.Debug(LOG_TAG, "Image runtime sync triggered.")
	err := h.SyncDataFromRedis()
	if err != nil {
		logger.Error(LOG_TAG, "Error while syncing data from redis ", err)
	}
}

func (h *ImageRuntimeCache) SyncDataFromRedis() error {
	versionFromRedis, err := h.ImageStore.GetHashSetVersion()
	if err != nil {
		logger.Error(LOG_TAG, "Error caught while getting hash set version from redis ", err)
		return err
	}
	if h.RuntimeMapVersion == -1 || h.RuntimeMapVersion != versionFromRedis {
		h.RuntimeMapVersion = versionFromRedis
		h.ImageStore.SyncDataFromRedis(h.ImageRuntimeMap)
		logger.Debug(LOG_TAG, "Updating image config map.")
		err := utils.CreateOrUpdateZkImageConfigMap(utils.GetCurrentNamespace(), common.ZkImageConfigMapName, h.ImageRuntimeMap)
		if err != nil {
			logger.Error(LOG_TAG, "Error while create/update image configmap ", err)
			return err
		}
		logger.Debug(LOG_TAG, "Updating process config map.")
		err = utils.CreateOrUpdateProcessInfoConfigMap(utils.GetCurrentNamespace(), common.ZkProcessConfigMapName, h.ImageRuntimeMap)
		if err != nil {
			logger.Error(LOG_TAG, "Error while create/update process info configmap ", err)
			return err
		}
	}
	return nil
}

func (h *ImageRuntimeCache) Init(config config.ZkOperatorConfig) {
	//init ImageStore
	h.ImageStore = GetNewRedisStore(config)
	h.RuntimeMapVersion = -1
	var err error
	h.ImageRuntimeMap, err = utils.GetSyncMapFromConfigMap(utils.GetCurrentNamespace(), common.ZkImageConfigMapName)
	logger.Debug(LOG_TAG, "ImageMap from configMap is ", h.ImageRuntimeMap)
	if err != nil {
		h.ImageRuntimeMap = &sync.Map{}
		logger.Error(LOG_TAG, "Error while reading image map from config Map ", err)
	}
	var duration = time.Duration(config.Instrumentation.PollingInterval) * time.Second
	h.ticker = zktick.GetNewTickerTask("images_sync", duration, h.periodicSync)
	h.ImageOverrideMap = &sync.Map{}
}

func (h *ImageRuntimeCache) getRuntimeForImage(imageID string) *common.ContainerRuntime {
	value, ok := h.ImageRuntimeMap.Load(imageID)
	if !ok {
		return nil
	}
	switch y := value.(type) {
	case *common.ContainerRuntime:
		return y
	default:
		return nil
	}
}

func (h *ImageRuntimeCache) GetContainerLanguage(container *corev1.Container, pod *corev1.Pod) common.ProgrammingLanguage {
	imageId := container.Image
	logger.Debug(LOG_TAG, "Image is ", imageId)
	runtime := h.getRuntimeForImage(imageId)
	if runtime == nil {
		return common.NotYetProcessed
	}
	languages := runtime.Languages
	if len(languages) > 0 {
		language := languages[0]
		logger.Debug(LOG_TAG, "found language ", language)
		if language == fmt.Sprintf("%v", common.JavaProgrammingLanguage) {
			return common.JavaProgrammingLanguage
		}
	}
	return common.UnknownLanguage
}

func (h *ImageRuntimeCache) CleanUpOnkill() error {
	logger.Debug(LOG_TAG, "Kill method in update orchestration.\n")
	h.ticker.Stop()
	return nil
}

func (h *ImageRuntimeCache) ProcessOverrideValues(values []v1alpha1.ImageOverride) error {
	logger.Debug(LOG_TAG, "Processing override values.")
	for _, imageOverride := range values {
		imageId := imageOverride.ImageID
		h.ImageOverrideMap.Store(imageId, imageOverride)
		logger.Debug(LOG_TAG, "Saving override value for imageId ", imageId)
	}
	return nil
}
