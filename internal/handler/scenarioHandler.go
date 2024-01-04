package handler

import (
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"

	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
	dbNames "github.com/zerok-ai/zk-utils-go/storage/redis/clientDBNames"
)

var scenarioLogTag = "ScenarioHandler"

type ScenarioHandler struct {
	VersionedStore *zkredis.VersionedStore[model.Scenario]
	config         config.ZkOperatorConfig
}

func (h *ScenarioHandler) Init(cfg config.ZkOperatorConfig) error {
	store, err := zkredis.GetVersionedStore[model.Scenario](&cfg.Redis, dbNames.ScenariosDBName, common.RedisSyncInterval)
	if err != nil {
		return err
	}
	h.VersionedStore = store
	h.config = cfg

	return nil
}

func (h *ScenarioHandler) CleanUpOnKill() error {
	logger.Debug(scenarioLogTag, "Kill method in scenario rules.")
	return nil
}

func (h *ScenarioHandler) IsHealthy() bool {
	return true
}
