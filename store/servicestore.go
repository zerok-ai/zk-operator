package store

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"github.com/zerok-ai/zk-utils-go/storage/redis/clientDBNames"
	"github.com/zerok-ai/zk-utils-go/storage/redis/config"
)

type ServiceStore struct {
	redisClient *redis.Client
}

func (s *ServiceStore) initialize() *ServiceStore {
	return s
}

func (s *ServiceStore) Close() {
	s.redisClient.Close()
}

func GetServiceStore(redisConfig config.RedisConfig) *ServiceStore {
	dbName := clientDBNames.ServiceListDBName
	_redisClient := config.GetRedisConnection(dbName, redisConfig)
	ss := ServiceStore{redisClient: _redisClient}
	serviceStore := ss.initialize()
	testConnection := serviceStore.redisClient.Ping(context.Background())
	if testConnection.Err() != nil {
		panic(testConnection.Err())
	} else {
		println("Connected to Redis")
	}
	return serviceStore
}

func (s *ServiceStore) GetServices() ([]string, error) {
	return s.redisClient.SMembers(context.Background(), utils.ServiceListRedisKey).Result()
}
