package storage

import (
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"sync"

	"github.com/go-redis/redis"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
)

type ImageStore struct {
	redisClient *redis.Client
	hashSetName string
}

func (zkRedis *ImageStore) GetHashSetVersion() (int64, error) {
	data, err := zkRedis.redisClient.Get(common.HashSetVersionKey).Int64()
	if err != nil {
		fmt.Printf("Error caught while getting hash set version from redis %v.\n ", err)
		return -1, err
	}
	return data, nil
}

func GetNewRedisStore(config config.ZkInjectorConfig) *ImageStore {
	_redisClient := utils.GetRedisClient(config, config.Redis.ImageDB)

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
		data, cursor, err = zkRedis.redisClient.HScan(common.HashSetName, cursor, "*", 10).Result()
		if err != nil {
			fmt.Printf("Error while scan from redis %v\n", err)
		}

		for i := 0; i < len(data); i += 2 {
			key := data[i]
			value := data[i+1]
			fmt.Printf("Key %v and va;ue %v.\n", key, value)
			serializedValue, err := utils.GetContainerRuntime(value)
			if err != nil {
				fmt.Printf("Error caught while serializing the value from redis %v for key %v.\n", err, key)
				continue
			}
			fmt.Println(key, value)
			//Updating data in the map.
			currMap.Store(key, serializedValue)
		}

		if cursor == 0 {
			break
		}
	}
}

func (zkRedis *ImageStore) GetString(key string) (*string, error) {
	output := zkRedis.redisClient.HGet(zkRedis.hashSetName, key)
	err := output.Err()
	if err != nil {
		return nil, err
	}
	value := output.Val()
	return &value, nil
}
