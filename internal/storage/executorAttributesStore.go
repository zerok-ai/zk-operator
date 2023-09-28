package storage

import (
	"github.com/redis/go-redis/v9"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
)

var LOG_TAG = "ExecutorAttributesStore"

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
	logger.Debug(LOG_TAG, "Uploading executor attributes to redis for executorVersionKey: "+executorVersionKey)
	_, err := zkRedis.redisClient.HMSet(ctx, executorVersionKey, executorAttributesMap).Result()
	if err != nil {
		return err
	}
	return nil
}

func (zkRedis *ExecutorAttributesStore) Close() {
	err := zkRedis.redisClient.Close()
	if err != nil {
		logger.Error(LOG_TAG, "Error while closing redis connection ", err)
		return
	}
	logger.Info(LOG_TAG, "Redis connection closed successfully")
}
