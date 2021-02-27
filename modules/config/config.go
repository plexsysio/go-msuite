package config

import (
	"io"
)

type Config interface {
	// For reading and writing from files
	io.Reader
	io.Writer

	// Print helpers
	String() string
	Pretty() string

	// Getters/Setters
	Get(key string, val interface{}) bool
	Set(key string, val interface{})

	// IsSet is helper used to check boolean value is set
	IsSet(key string) bool
}
