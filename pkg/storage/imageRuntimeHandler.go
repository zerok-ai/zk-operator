package storage

import (
	"fmt"
	"sync"

	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/pkg/common"
	"github.com/zerok-ai/zk-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
)

type ImageRuntimeHandler struct {
	ImageRuntimeMap   *sync.Map
	RuntimeMapVersion int64
	ImageStore        Store
}

func (h *ImageRuntimeHandler) SyncDataFromRedis() error {
	versionFromRedis, err := h.ImageStore.GetHashSetVersion()
	if err != nil {
		fmt.Printf("Error caught while getting hash set version from redis %v.\n", err)
		return err
	}
	if h.RuntimeMapVersion == -1 || h.RuntimeMapVersion != versionFromRedis {
		h.RuntimeMapVersion = versionFromRedis
		h.ImageStore.SyncDataFromRedis(h.ImageRuntimeMap)
	}
	return nil
}

func (h *ImageRuntimeHandler) Init(config config.ZkInjectorConfig) {
	//init ImageStore
	h.ImageStore = GetNewRedisStore(config)
	h.RuntimeMapVersion = -1
	h.ImageRuntimeMap = &sync.Map{}
}

func (h *ImageRuntimeHandler) getRuntimeForImage(imageID string) *common.ContainerRuntime {
	value, ok := h.ImageRuntimeMap.Load(imageID)
	if !ok {
		return nil
	}
	switch y := value.(type) {
	case *common.ContainerRuntime:
		fmt.Println("mk: Getting data for image id ", imageID, utils.ToJsonString(y))
		return y
	default:
		return nil
	}
}

func (h *ImageRuntimeHandler) GetContainerLanguage(container *corev1.Container, pod *corev1.Pod) common.ProgrammingLanguage {
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
