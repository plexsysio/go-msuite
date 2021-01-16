package locker

import (
	"github.com/StreamSpace/ss-store"
	"time"
)

const (
	// lock acquire timeout
	// DefaultTimeout lock acquire timeout
	DefaultTimeout = time.Millisecond * 1000
)

type (
	// Locker Interface is the base functionality that any locker handler
	// should implement in order to become valid handler
	Locker interface {
		Close() error
		Lock(doc store.Item) (func() error, error)
		TryLock(doc store.Item, t time.Duration) (func() error, error)
	}
)
