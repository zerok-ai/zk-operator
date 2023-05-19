package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/zerok-ai/zk-operator/internal/utils"

	"github.com/go-redis/redis"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
)

type RedisStore struct {
	redisClient *redis.Client
	hashSetName string
}

func (zkRedis *RedisStore) GetHashSetVersion() (int64, error) {
	data, err := zkRedis.redisClient.Get(common.HashSetVersionKey).Int64()
	if err != nil {
		fmt.Printf("Error caught while getting hash set version from redis %v.\n ", err)
		return -1, err
	}
	return data, nil
}

func GetNewRedisStore(config config.ZkInjectorConfig) Store {
	redisConfig := config.Redis
	readTimeout := time.Duration(redisConfig.ReadTimeout) * time.Second
	addr := fmt.Sprint(redisConfig.Host, ":", redisConfig.Port)
	fmt.Printf("Address for redis is %v.\n", addr)
	_redisClient := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    "",
		DB:          redisConfig.DB,
		ReadTimeout: readTimeout,
	})

	imgRedis := &RedisStore{
		redisClient: _redisClient,
		hashSetName: common.HashSetName,
	}

	return imgRedis
}

func (zkRedis *RedisStore) SyncDataFromRedis(currMap *sync.Map) {
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

func (zkRedis *RedisStore) GetString(key string) (*string, error) {
	output := zkRedis.redisClient.HGet(zkRedis.hashSetName, key)
	err := output.Err()
	if err != nil {
		return nil, err
	}
	value := output.Val()
	return &value, nil
}
