package storage

import (
	"sync"
)

type Store interface {
	GetHashSetVersion() (int64, error)
	LoadAllData() (*sync.Map, error)
	GetString(key string) (*string, error)
}
