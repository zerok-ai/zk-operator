package internal

type ZkOperatorModule interface {
	CleanUpOnkill() error
}
