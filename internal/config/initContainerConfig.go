package config

import (
	logger "github.com/zerok-ai/zk-utils-go/logs"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	"time"
)

var LOG_TAG = "AppInitContainerData"

type AppInitContainerData struct {
	image  string
	tag    string
	ticker *zktick.TickerTask
}

func (h *AppInitContainerData) Init(cfg ZkOperatorConfig) {
	//Creating a timer for periodic app init container sync.
	var duration = time.Duration(cfg.ScenarioSync.PollingInterval) * time.Second
	h.ticker = zktick.GetNewTickerTask("scenario_sync", duration, h.periodicSync)
}

func (h *AppInitContainerData) StartPeriodicSync() {
	h.updateInitContainerConfig()
	h.ticker.Start()
}

func (h *AppInitContainerData) periodicSync() {
	logger.Debug(LOG_TAG, "Init container data sync triggered.")
	h.updateInitContainerConfig()
}

func (h *AppInitContainerData) updateInitContainerConfig() {

}
