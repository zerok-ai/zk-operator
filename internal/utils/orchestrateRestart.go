package utils

import (
	"github.com/zerok-ai/zk-operator/internal/storage"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	ticker "github.com/zerok-ai/zk-utils-go/ticker"
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
	zklogger.Debug(LOG_TAG, "Restarting marked namespaces if needed")
	err := RestartMarkedNamespacesIfNeeded(false, orchestrateRestart.imageRuntimeCache)
	if err != nil {
		zklogger.Error(LOG_TAG, "Error while restarting marked namespaces if needed ", err)
	}
}
