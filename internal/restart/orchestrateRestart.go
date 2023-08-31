package restart

import (
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/storage"
	"github.com/zerok-ai/zk-operator/internal/utils"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	ticker "github.com/zerok-ai/zk-utils-go/ticker"
	"k8s.io/api/core/v1"
	"time"
)

type OrchestrateRestart struct {
	imageRuntimeCache *storage.ImageRuntimeCache
	Ticker            *ticker.TickerTask
}

var orchestrateRestart OrchestrateRestart

func NewOrchestrateRestart(imageRuntimeCache *storage.ImageRuntimeCache) *OrchestrateRestart {
	tickerTask := ticker.GetNewTickerTask("orchestrate_restart_ticker", time.Duration(2)*time.Minute, restartNonOrchestratedPodsIfNeeded)

	orchestrateRestart = OrchestrateRestart{
		imageRuntimeCache: imageRuntimeCache,
		Ticker:            tickerTask,
	}

	return &orchestrateRestart
}

func restartNonOrchestratedPodsIfNeeded() {
	zklogger.Debug(utils.LOG_TAG, "Restarting marked namespaces if needed")
	err := RestartMarkedNamespacesIfNeeded(false, orchestrateRestart.imageRuntimeCache)
	if err != nil {
		zklogger.Error(utils.LOG_TAG, "Error while restarting marked namespaces if needed ", err)
	}
}

func RestartMarkedNamespacesIfNeeded(orchestratedPods bool, imageRuntimeCache *storage.ImageRuntimeCache) error {
	zklogger.Debug(utils.LOG_TAG, "In restart marked namespaces with orchestrated pods flag as ", orchestratedPods)
	namespaces, err := utils.GetAllMarkedNamespaces()
	zklogger.Debug(utils.LOG_TAG, "All marked namespaces are ", namespaces)

	if err != nil || namespaces == nil {
		zklogger.Error(utils.LOG_TAG, "In restart marked namespaces, error caught while getting all marked namespaces ", err)
		return err
	}

	for _, namespace := range namespaces.Items {

		zklogger.Debug(utils.LOG_TAG, " Checking for namespace ", namespace.ObjectMeta.Name)

		var pods []v1.Pod

		if orchestratedPods {
			pods, err = utils.GetOrchestratedPods(namespace.ObjectMeta.Name)
			if err != nil {
				zklogger.Error(utils.LOG_TAG, "Error caught while getting all non orchestrated pods ", err)
				return err
			}
		} else {
			pods, err = utils.GetNotOrchestratedPods(namespace.ObjectMeta.Name)
			if err != nil {
				zklogger.Error(utils.LOG_TAG, "Error caught while getting all non orchestrated pods ", err)
				return err
			}

			if imageRuntimeCache != nil {

				podsToOrchestrate := make([]v1.Pod, 0)
				for _, pod := range pods {
					containers := pod.Spec.Containers
					for index := range containers {
						container := &pod.Spec.Containers[index]
						language := imageRuntimeCache.GetContainerLanguage(container, nil)
						if language == common.JavaProgrammingLanguage {
							podsToOrchestrate = append(podsToOrchestrate, pod)
							break
						}
					}
				}
			}
		}

		zklogger.Debug(utils.LOG_TAG, " Non orchestrated pods for namespace ", namespace.ObjectMeta.Name, " are ", pods)

		workLoads, err := utils.GetWorkloadsForPods(pods)
		if err != nil {
			return err
		}

		zklogger.Debug(utils.LOG_TAG, " Workloads for pods for namespace ", namespace.ObjectMeta.Name, " are ", workLoads)

		for workLoad := range workLoads {
			restart, err := utils.HasRestartLabel(namespace.ObjectMeta.Name, workLoad.WorkLoadType, workLoad.Name, common.ZkAutoRestartKey, common.ZkAutoRestartValue)
			if err != nil {
				zklogger.Error(utils.LOG_TAG, "Error caught while checking if workload ", workLoad, " has restart label ", err)
			}
			if !restart {
				zklogger.Debug(utils.LOG_TAG, "Workload ", workLoad.Name, " and type ", workLoad.WorkLoadType, " does not have restart label, skipping")
				continue
			}
			switch workLoad.WorkLoadType {
			case utils.DEPLYOMENT:
				err = utils.RestartDeployment(namespace.ObjectMeta.Name, workLoad.Name)
			case utils.STATEFULSET:
				err = utils.RestartStatefulSet(namespace.ObjectMeta.Name, workLoad.Name)
			case utils.DAEMONSET:
				err = utils.RestartDaemonSet(namespace.ObjectMeta.Name, workLoad.Name)
			default:
				zklogger.Error(utils.LOG_TAG, "Unknown workload type ", workLoad.WorkLoadType)

			}
			if err != nil {
				zklogger.Error(utils.LOG_TAG, "Error caught while restarting workload name ", workLoad, " with type ", workLoad.WorkLoadType, " with error ", err)
				return err
			}
		}
	}
	return nil
}
