package sync

import (
	"fmt"
	"time"

	"github.com/zerok-ai/operator/internal/config"
	"github.com/zerok-ai/operator/pkg/storage"
	"github.com/zerok-ai/operator/pkg/zkclient"
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

func restartMarkedNamespacesIfNeeded() error {
	namespaces, err := zkclient.GetAllMarkedNamespaces()

	if err != nil || namespaces == nil {
		fmt.Printf("In restart marked namespaces, error caught while getting all marked namespaces %v.\n", err)
		return err
	}

	for _, namespace := range namespaces.Items {
		deployments := make(map[string]bool)
		pods, err := zkclient.GetNotOrchestratedPods(namespace.ObjectMeta.Name)
		if err != nil {
			fmt.Printf("Error caught while getting all non orchestrated pods %v.\n", err)
			return err
		}
		for _, pod := range pods {
			deploymentName, err := zkclient.GetDeploymentForPods(&pod)
			if err != nil {
				fmt.Printf("Error caught while getting all deployment for pod %v with error %v.\n", deploymentName, err)
				return err
			}
			deployments[deploymentName] = true
		}
		for deploymentName := range deployments {
			err = zkclient.RestartDeployment(namespace.ObjectMeta.Name, deploymentName)
			if err != nil {
				fmt.Printf("Error caught while restaring deployment name %v with error %v.\n", deploymentName, err)
				return err
			}
		}
	}
	return nil
}
