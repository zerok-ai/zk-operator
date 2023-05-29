package internal

type Zkmodule interface {
	CleanUpOnkill() error
}
