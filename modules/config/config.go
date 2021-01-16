package config

import (
	"io"
)

type Config interface {
	io.Reader
	// Filename if used
	FileName(string) string
	// Print helpers
	String() string
	Pretty() string

	// Getters/Setters
	Get(key string, val interface{}) bool
	Set(key string, val interface{})
}
