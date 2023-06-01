package storage

import (
	"fmt"
	common "github.com/zerok-ai/zk-operator/internal/common"
	utils "github.com/zerok-ai/zk-operator/internal/utils"
	"sync"
	"time"

	"github.com/zerok-ai/zk-operator/internal/config"
	corev1 "k8s.io/api/core/v1"
)

type ImageRuntimeCache struct {
	ImageRuntimeMap   *sync.Map
	RuntimeMapVersion int64
	ImageStore        *ImageStore
	ticker            *time.Ticker
}

func (h *ImageRuntimeCache) StartPeriodicSync() {
	//Sync first time on pod start
	err := h.SyncDataFromRedis()
	if err != nil {
		fmt.Printf("Error while syncing data from redis %v.\n", err)
	}
	for range h.ticker.C {
		fmt.Println("Sync triggered.")
		err = h.SyncDataFromRedis()
		if err != nil {
			fmt.Printf("Error while syncing data from redis %v.\n", err)
		}
		err = utils.RestartMarkedNamespacesIfNeeded()
		if err != nil {
			fmt.Printf("Error while restarting marked namespaces if needed %v.\n", err)
		}
	}
}

func (h *ImageRuntimeCache) SyncDataFromRedis() error {
	versionFromRedis, err := h.ImageStore.GetHashSetVersion()
	if err != nil {
		fmt.Printf("Error caught while getting hash set version from redis %v.\n", err)
		return err
	}
	if h.RuntimeMapVersion == -1 || h.RuntimeMapVersion != versionFromRedis {
		h.RuntimeMapVersion = versionFromRedis
		h.ImageStore.SyncDataFromRedis(h.ImageRuntimeMap)
		fmt.Println("Updating config map.")
		err := utils.CreateOrUpdateConfigMap(utils.GetCurrentNamespace(), common.ZkConfigMapName, h.ImageRuntimeMap)
		if err != nil {
			fmt.Printf("Error while create/update confimap %v.\n", err)
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
	h.ImageRuntimeMap, err = utils.GetDataFromConfigMap(utils.GetCurrentNamespace(), common.ZkConfigMapName)
	fmt.Printf("ImageMap from configMap is %v.\n", h.ImageRuntimeMap)
	if err != nil {
		h.ImageRuntimeMap = &sync.Map{}
		fmt.Printf("Error while reading image map from config Map %v.\n", err)
	}
	var duration = time.Duration(config.Redis.PollingInterval) * time.Second
	h.ticker = time.NewTicker(duration)
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
	fmt.Printf("Image is %v.\n", imageId)
	runtime := h.getRuntimeForImage(imageId)
	if runtime == nil {
		return common.NotYetProcessed
	}
	languages := runtime.Languages
	if len(languages) > 0 {
		language := languages[0]
		fmt.Println("found language ", language)
		if language == fmt.Sprintf("%v", common.JavaProgrammingLanguage) {
			return common.JavaProgrammingLanguage
		}
	}
	return common.UknownLanguage
}

func (h *ImageRuntimeCache) CleanUpOnkill() error {
	fmt.Printf("Kill method in update orchestration.\n")
	h.ticker.Stop()
	return nil
}
