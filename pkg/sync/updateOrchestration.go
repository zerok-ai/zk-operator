package sync

import (
	"fmt"
	"github.com/zerok-ai/zk-operator/pkg/utils"
	"time"

	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/pkg/storage"
	v1 "k8s.io/api/core/v1"
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
	namespaces, err := utils.GetAllMarkedNamespaces()

	if err != nil || namespaces == nil {
		fmt.Printf("In restart marked namespaces, error caught while getting all marked namespaces %v.\n", err)
		return err
	}

	for _, namespace := range namespaces.Items {

		pods, err := utils.GetNotOrchestratedPods(namespace.ObjectMeta.Name)
		if err != nil {
			fmt.Printf("Error caught while getting all non orchestrated pods %v.\n", err)
			return err
		}

		deployments, err := getDeploymentsForPods(pods)
		if err != nil {
			return err
		}

		for deploymentName := range deployments {
			err = utils.RestartDeployment(namespace.ObjectMeta.Name, deploymentName)
			if err != nil {
				fmt.Printf("Error caught while restaring deployment name %v with error %v.\n", deploymentName, err)
				return err
			}
		}
	}
	return nil
}

func getDeploymentsForPods(pods []v1.Pod) (map[string]bool, error) {
	deployments := make(map[string]bool)
	for _, pod := range pods {
		deploymentName, err := utils.GetDeploymentForPods(&pod)
		if err != nil {
			fmt.Printf("Error caught while getting all deployment for pod %v with error %v.\n", deploymentName, err)
			return deployments, err
		}
		deployments[deploymentName] = true
	}
	return deployments, nil
}
