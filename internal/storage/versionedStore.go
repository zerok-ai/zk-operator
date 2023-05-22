package storage

import (
	"github.com/go-redis/redis"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
)

type VersionedStore struct {
	redisClient        *redis.Client
	versionHashSetName string
}

func GetVersionedStore(config config.ZkInjectorConfig) *VersionedStore {

	_redisClient := utils.GetRedisClient(config, config.Redis.VersionDB)

	versionStore := &VersionedStore{
		redisClient:        _redisClient,
		versionHashSetName: "zk_value_version",
	}

	return versionStore
}

func (zkRedis VersionedStore) SetValue(key string, value string) error {
	rdb := zkRedis.redisClient

	// get the old value
	oldVal := rdb.Get(key)
	if err := oldVal.Err(); err != nil {
		return err
	}
	oldString := oldVal.Val()

	// return if old value is same as new value
	if oldString == value {
		return nil
	}

	// Create a Redis transaction: this doesn't support rollback
	tx := rdb.TxPipeline()

	// set value and version
	tx.Set(key, value, 0)
	tx.HIncrBy(zkRedis.versionHashSetName, key, 1)

	// Execute the transaction
	if _, err := tx.Exec(); err != nil {
		return err
	}

	return nil
}

func (zkRedis VersionedStore) GetAllVersions() (map[string]string, error) {
	rdb := zkRedis.redisClient

	// get the old value
	versions := rdb.HGetAll(zkRedis.versionHashSetName)
	return versions.Val(), versions.Err()
}

func (zkRedis VersionedStore) GetValue(key string) (string, error) {
	rdb := zkRedis.redisClient

	// get the old value
	value := rdb.Get(key)
	return value.Val(), value.Err()
}

func (zkRedis VersionedStore) GetValuesForKeys(keys []string) ([]interface{}, error) {
	rdb := zkRedis.redisClient

	// get the values
	return rdb.MGet(keys...).Result()
}

func (zkRedis VersionedStore) GetVersion(key string) (string, error) {
	rdb := zkRedis.redisClient

	// get the old value
	version := rdb.HGet(zkRedis.versionHashSetName, key)
	return version.Val(), version.Err()
}

func (zkRedis VersionedStore) Delete(key string) error {
	rdb := zkRedis.redisClient

	// create a transaction
	tx := rdb.TxPipeline()

	// delete version
	tx.HDel(zkRedis.versionHashSetName, key)
	tx.Del(key)

	// Execute the transaction
	if _, err := tx.Exec(); err != nil {
		return err
	}
	return nil
}

func (zkRedis VersionedStore) Length() (int64, error) {
	// get the number of hash key-value pairs
	return zkRedis.redisClient.HLen(zkRedis.versionHashSetName).Result()
}
