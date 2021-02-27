package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	defer fp.Close()
	_, err = io.Copy(fp, r)
	return err
}

func ReadFromFile(w io.Writer, f string) error {
	fp, err := os.Open(f)
	if err != nil {
		return err
	}
	defer fp.Close()
	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}
	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n < len(buf) {
		return errors.New("Unable to write entire config")
	}
	return nil
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
