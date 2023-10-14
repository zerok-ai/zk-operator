package storage

import (
	"github.com/redis/go-redis/v9"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/storage/redis/clientDBNames"
)

var LOG_TAG_ATTR_STORE = "ExecutorAttributesStore"

type ExecutorAttributesStore struct {
	redisClient *redis.Client
}

const (
	LatestVersionKey = "latest_version"
)

func GetExecutorAttributesRedisStore(config config.ZkOperatorConfig) *ExecutorAttributesStore {
	_redisClient := utils.GetRedisClient(clientDBNames.ExecutorAttrDBName, config.Redis)

	executorAttributesStore := &ExecutorAttributesStore{
		redisClient: _redisClient,
	}

	return executorAttributesStore
}

func (zkRedis *ExecutorAttributesStore) UploadExecutorAttributes(executorVersionKey string, executorAttributesMap map[string]interface{}) error {
	logger.Debug(LOG_TAG_ATTR_STORE, "Uploading executor attributes to redis for executorVersionKey: "+executorVersionKey)
	logger.Debug(LOG_TAG_ATTR_STORE, "Key and value are ")
	for key, value := range executorAttributesMap {
		logger.Debug(LOG_TAG_ATTR_STORE, key, value)
	}
	_, err := zkRedis.redisClient.HMSet(ctx, executorVersionKey, executorAttributesMap).Result()
	if err != nil {
		return err
	}
	return nil
}

func (zkRedis *ExecutorAttributesStore) UpdateLastSyncTime(latestVersion string) error {
	logger.Debug(LOG_TAG_ATTR_STORE, "Updating latest version to redis")
	_, err := zkRedis.redisClient.Set(ctx, LatestVersionKey, latestVersion, 0).Result()
	if err != nil {
		return err
	}
	return nil
}

func (zkRedis *ExecutorAttributesStore) GetLastSyncVersion() string {
	logger.Debug(LOG_TAG_ATTR_STORE, "Getting latest version from redis")
	latestVersion, err := zkRedis.redisClient.Get(ctx, LatestVersionKey).Result()
	if err != nil {
		logger.Error(LOG_TAG_ATTR_STORE, "Error in getting latest version from redis ", err)
		return "0"
	}
	return latestVersion
}

func (zkRedis *ExecutorAttributesStore) Close() {
	err := zkRedis.redisClient.Close()
	if err != nil {
		logger.Error(LOG_TAG_ATTR_STORE, "Error while closing redis connection ", err)
		return
	}
	logger.Info(LOG_TAG_ATTR_STORE, "Redis connection closed successfully")
}
