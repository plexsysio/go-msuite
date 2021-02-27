package jsonConf

import (
	"encoding/json"
	"io"

	"github.com/aloknerurkar/go-msuite/utils"
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

func DefaultConfig() *JsonConfig {
	var conf = make(JsonConfig)
	return &conf
}

func FromFile(f string) (*JsonConfig, error) {
	j := &JsonConfig{}
	err := utils.ReadFromFile(j, f)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (j *JsonConfig) String() string {
	buf, err := json.Marshal(j)
	if err != nil {
		return "INVALID_CONFIG"
	}
	return string(buf)
}

func (j *JsonConfig) Read(p []byte) (int, error) {
	buf, err := json.MarshalIndent(j, "", "\t")
	if err != nil {
		return 0, err
	}
	copy(p, buf)
	if len(buf) > len(p) {
		return len(p), nil
	}
	return len(buf), io.EOF
}

func (j *JsonConfig) Write(p []byte) (int, error) {
	err := json.Unmarshal(p, j)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (j *JsonConfig) Pretty() string {
	buf, err := json.MarshalIndent(j, "", "\t")
	if err != nil {
		return "INVALID_CONFIG"
	}
	return string(buf)
}
