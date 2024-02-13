package store

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/zerok-ai/zk-utils-go/storage/redis/clientDBNames"
	"github.com/zerok-ai/zk-utils-go/storage/redis/config"
)

type ServiceStore struct {
	redisClient *redis.Client
}

func (t ServiceStore) initialize() *ServiceStore {
	return &t
}

func (t ServiceStore) Close() {
	t.redisClient.Close()
}

func GetServiceStore(redisConfig config.RedisConfig) *ServiceStore {
	dbName := clientDBNames.ServiceListDBName
	_redisClient := config.GetRedisConnection(dbName, redisConfig)
	serviceStore := ServiceStore{redisClient: _redisClient}.initialize()
	return serviceStore
}

func (t ServiceStore) GetServices() ([]string, error) {
	return t.redisClient.SMembers(context.Background(), "services").Result()
}
