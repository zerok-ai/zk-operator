package sync

import (
	"fmt"
	"time"
	"github.com/zerok-ai/operator/internal/config"
	"github.com/zerok-ai/operator/pkg/storage"
)

func UpdateOrchestration(imageRuntimeHandler *storage.ImageRuntimeHandler, cfg config.ZkInjectorConfig) {
	//Sync first time on pod start
	imageRuntimeHandler.SyncDataFromRedis()

	//Creating a timer for periodic sync
	var duration = time.Duration(cfg.Redis.PollingInterval) * time.Second
	ticker := time.NewTicker(duration)
	for range ticker.C {
		fmt.Println("Sync triggered.")
		imageRuntimeHandler.SyncDataFromRedis()
		restartMarkedNamespacesIfNeeded()
	}
}

func restartMarkedNamespacesIfNeeded() {
	//TODO: Get List of Deployments which need to restarted only if the auto restart flag is present.
}
