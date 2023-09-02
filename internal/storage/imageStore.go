package storage

import (
	"context"
	"github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	redisConfig "github.com/zerok-ai/zk-utils-go/storage/redis/config"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
)

var ctx = context.Background()
var LOG_TAG_STORE = "ImageStore"

type ImageStore struct {
	redisClient *redis.Client
	hashSetName string
}

func (zkRedis *ImageStore) GetHashSetVersion() (int64, error) {
	data, err := zkRedis.redisClient.Get(ctx, common.HashSetVersionKey).Int64()
	if err != nil {
		logger.Error(LOG_TAG_STORE, "Error caught while getting hash set version from redis ", err)
		return -1, err
	}
	return data, nil
}

func GetNewRedisStore(config config.ZkOperatorConfig) *ImageStore {
	_redisClient := redisConfig.GetRedisConnection(common.RedisImageDbName, config.Redis)

	imgRedis := &ImageStore{
		redisClient: _redisClient,
		hashSetName: common.HashSetName,
	}

	return imgRedis
}

func (zkRedis *ImageStore) SyncDataFromRedis(currMap *sync.Map) {
	var cursor uint64
	var data []string

	for {
		var err error
		//Getting 10 fields at once.
		data, cursor, err = zkRedis.redisClient.HScan(ctx, common.HashSetName, cursor, "*", 10).Result()
		if err != nil {
			logger.Error(LOG_TAG_STORE, "Error while scan from redis ", err)
		}

		for i := 0; i < len(data); i += 2 {
			key := data[i]
			value := data[i+1]
			//logger.Debug(LOG_TAG_STORE, "Key and value are ", key, value)
			serializedValue, err := utils.GetContainerRuntime(value)
			if err != nil {
				logger.Error(LOG_TAG_STORE, "Error caught while serializing the value from redis ", err, " for key ", key)
				continue
			}
			//logger.Debug(LOG_TAG_STORE, key, value)
			//Updating data in the map.
			currMap.Store(key, serializedValue)
		}

		if cursor == 0 {
			break
		}
	}
}

func (zkRedis *ImageStore) GetString(key string) (*string, error) {
	output := zkRedis.redisClient.HGet(ctx, zkRedis.hashSetName, key)
	err := output.Err()
	if err != nil {
		return nil, err
	}
	value := output.Val()
	return &value, nil
}
