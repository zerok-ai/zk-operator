package storage

import "sync"

type Store interface {
	GetHashSetVersion() (int64, error)
	SyncDataFromRedis(currMap *sync.Map)
	GetString(key string) (*string, error)
}
