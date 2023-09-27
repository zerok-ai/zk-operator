package storage

import (
	"github.com/redis/go-redis/v9"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
)

type ExecutorAttributesStore struct {
	redisClient *redis.Client
}

func GetExecutorAttributesRedisStore(config config.ZkOperatorConfig) *ExecutorAttributesStore {
	_redisClient := utils.GetRedisClient(common.ExecutorAttrDbName, config.Redis)

	executorAttributesStore := &ExecutorAttributesStore{
		redisClient: _redisClient,
	}

	return executorAttributesStore
}

func (zkRedis *ExecutorAttributesStore) UploadExecutorAttributes(executorVersionKey string, executorAttributesMap map[string]interface{}) error {
	_, err := zkRedis.redisClient.HMSet(ctx, executorVersionKey, executorAttributesMap).Result()
	if err != nil {
		return err
	}
	return nil
}

func (zkRedis *ExecutorAttributesStore) Close() {
	err := zkRedis.redisClient.Close()
	if err != nil {
		logger.Error("Error while closing redis connection ", err)
		return
	}
	logger.Info("Redis connection closed successfully")
}
