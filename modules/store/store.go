package store

import (
	"os"
)

const (
	// SortNatural use natural order
	SortNatural Sort = iota
	// SortCreatedDesc created newest to oldest
	SortCreatedDesc
	// SortCreatedAsc created oldest to newset
	SortCreatedAsc
	// SortUpdatedDesc updated newest to oldest
	SortUpdatedDesc
	// SortUpdatedAsc updated oldest to newset
	SortUpdatedAsc
)

type (
	Sort int

	Items []Item

	Item interface {
		GetNamespace() string
		GetId() string
	}

	Exportable interface {
		Marshal() ([]byte, error)
		Unmarshal([]byte) error
	}

	FileItemSetter interface {
		SetFp(*os.File)
	}

	TimeTracker interface {
		SetCreated(t int64)
		GetCreated() int64
		SetUpdated(t int64)
		GetUpdated() int64
	}

	IdSetter interface {
		SetId(string)
	}

	ListOpt struct {
		Page    int64
		Limit   int64
		Sort    Sort
		Version int64
	}

	Store interface {
		Create(Item) error
		Update(Item) error
		Delete(Item) error
		Read(Item) error
		List(Items, ListOpt) (int, error)
	}
)
