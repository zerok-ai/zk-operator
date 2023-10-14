package internal

type ZkOperatorModule interface {
	CleanUpOnKill() error
	IsHealthy() bool
}
