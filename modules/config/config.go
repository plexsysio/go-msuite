package config

import (
	"io"
)

type Config interface {
	io.Reader
	// Print helpers
	String() string
	Pretty() string

	// Getters/Setters
	Get(key string, val interface{}) bool
	Set(key string, val interface{})
}

func FromFile(filepath string) (Config, error) {
	return nil, nil
}
