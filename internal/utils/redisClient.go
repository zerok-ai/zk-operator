package utils

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zerok-ai/zk-utils-go/storage/redis/config"
	"time"
)

func GetRedisConnection(dbName string, redisConfig config.RedisConfig) *redis.Client {
	readTimeout := time.Duration(redisConfig.ReadTimeout) * time.Second
	return redis.NewClient(&redis.Options{
		Addr:        fmt.Sprint(redisConfig.Host, ":", redisConfig.Port),
		Password:    redisConfig.Password,
		DB:          redisConfig.DBs[dbName],
		ReadTimeout: readTimeout,
	})
}

func GetRedisClient(redisDBName string, redisConfig config.RedisConfig) *redis.Client {
	return GetRedisConnection(redisDBName, redisConfig)
}