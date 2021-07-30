package jsonConf

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/plexsysio/go-msuite/utils"
)

type JsonConfig map[string]interface{}

func (j *JsonConfig) Get(key string, val interface{}) bool {
	_, ok := (*j)[key]
	if !ok {
		return false
	}
	jsonString, err := json.Marshal((*j)[key])
	if err != nil {
		return false
	}
	if err := json.Unmarshal(jsonString, val); err != nil {
		return false
	}
	return true
}

func (j *JsonConfig) Set(key string, val interface{}) {
	(*j)[key] = val
}

func (j *JsonConfig) IsSet(key string) bool {
	val, ok := (*j)[key]
	return ok && val.(bool)
}

func (j *JsonConfig) Exists(key string) bool {
	_, ok := (*j)[key]
	return ok
}

func DefaultConfig() *JsonConfig {
	var conf = make(JsonConfig)
	return &conf
}

func FromFile(f string) (*JsonConfig, error) {
	j := &JsonConfig{}
	writer := j.Writer()
	err := utils.ReadFromFile(writer, f)
	if err != nil {
		return nil, err
	}
	return j, writer.Close()
}

func (j *JsonConfig) String() string {
	buf, err := json.Marshal(j)
	if err != nil {
		return "INVALID_CONFIG"
	}
	return string(buf)
}

func (j *JsonConfig) Reader() (io.Reader, error) {
	buf, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(buf), nil
}

type writeCloser struct {
	*bytes.Buffer
	conf *JsonConfig
}

func (w *writeCloser) Close() error {
	err := json.Unmarshal(w.Bytes(), w.conf)
	if err != nil {
		return err
	}
	w.Reset()
	return nil
}

func (j *JsonConfig) Writer() io.WriteCloser {
	return &writeCloser{Buffer: bytes.NewBuffer(nil), conf: j}
}

func (j *JsonConfig) Pretty() string {
	buf, err := json.MarshalIndent(j, "", "\t")
	if err != nil {
		return "INVALID_CONFIG"
	}
	return string(buf)
}
