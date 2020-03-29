package locker

import (
	storeItem "github.com/aloknerurkar/go-msuite/modules/store"
	"time"
)

const (
	// lock acquire timeout
	// DefaultTimeout lock acquire timeout
	DefaultTimeout = time.Millisecond * 30000
)

type (
	// Locker Interface is the base functionality that any locker handler
	// should implement in order to become valid handler
	Locker interface {
		Close() error
		Lock(doc storeItem.Item) (func() error, error)
		TryLock(doc storeItem.Item, t time.Duration) (func() error, error)
	}
)
