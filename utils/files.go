package utils

import (
	"io"
	"os"
)

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func WriteToFile(r io.Reader, newFile string) error {
	fp, err := os.Create(newFile)
	if err != nil {
		return err
	}
	_, err = io.Copy(fp, r)
	return err
}

func MkdirIfNotExists(path string) error {
	if !Exists(path) {
		err := os.MkdirAll(path, 0775)
		if err != nil {
			return err
		}
	}
	return nil
}
