package internal

type ZkModule interface {
	CleanUpOnkill() error
}
