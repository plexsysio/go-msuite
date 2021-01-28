package jsonConf

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/aloknerurkar/go-msuite/utils"
	logger "github.com/ipfs/go-log/v2"
)

var log = logger.Logger("jsonConfig")

const (
	ApiPort     = "4341"
	SwarmPort   = "4342"
	GatewayPort = "4343"
)

type JsonConfig map[string]interface{}

func (j *JsonConfig) Get(key string, val interface{}) bool {
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

func DefaultConfig() *JsonConfig {
	var conf = make(JsonConfig)
	conf["SwarmPort"] = SwarmPort
	conf["APIPort"] = ApiPort
	conf["GatewayPort"] = GatewayPort
	conf["SwarmAddrs"] = []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", SwarmPort),
		fmt.Sprintf("/ip6/::/tcp/%s", SwarmPort),
	}
	conf["ReproviderInterval"] = "12h"
	conf["Store"] = "bolt"
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
