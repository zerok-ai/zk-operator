package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/pkg/utils"

	"github.com/go-redis/redis"
)

const (
	defaultExpiry     time.Duration = time.Hour * 24 * 30
	hashSetName       string        = "zk_img_proc_map"
	hashSetVersionKey string        = "zk_img_proc_version"
)

type RedisStore struct {
	redisClient *redis.Client
	hashSetName string
}

func (zkRedis *RedisStore) GetHashSetVersion() (int64, error) {
	data, err := zkRedis.redisClient.Get(hashSetVersionKey).Int64()
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
		hashSetName: hashSetName,
	}

	return imgRedis
}

func (zkRedis *RedisStore) LoadAllData() (*sync.Map, error) {
	var cursor uint64
	var data []string

	tempMap := &sync.Map{}

	for {
		var err error
		//Getting 10 fields at once.
		data, cursor, err = zkRedis.redisClient.HScan(hashSetName, cursor, "*", 10).Result()
		if err != nil {
			fmt.Printf("Error while scan from redis %v\n", err)
			return nil, err
		}

		for i := 0; i < len(data); i += 2 {
			key := data[i]
			value := data[i+1]
			serializedValue, err := utils.FromJsonString(value)
			if err != nil {
				fmt.Printf("Error caught while serializing the value from redis %v.\n", err)
				//Abandoning load from redis, since data doesn't match expected type.
				return nil, err
			}
			fmt.Println(key, value)
			//Saving data to a tempe Map, and this map will be written to main Map.
			tempMap.Store(key, serializedValue)
		}

		if cursor == 0 {
			break
		}
	}

	return tempMap, nil
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
