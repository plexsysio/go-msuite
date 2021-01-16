package jsonConf

import (
	"encoding/json"
	"fmt"

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

// Functions for Store and Item interfaces
func (j *JsonConfig) GetNamespace() string {
	return "msuiteConfig"
}

func (j *JsonConfig) GetId() string {
	return "1"
}

func (j *JsonConfig) Marshal() ([]byte, error) {
	return json.Marshal(j)
}

func (j *JsonConfig) Unmarshal(buf []byte) error {
	return json.Unmarshal(buf, j)
}

func (j *JsonConfig) String() string {
	buf, err := json.Marshal(j)
	if err != nil {
		return "INVALID_CONFIG"
	}
	return string(buf)
}

func (j *JsonConfig) Json() []byte {
	buf, err := json.MarshalIndent(j, "", "\t")
	if err != nil {
		return []byte("INVALID_CONFIG")
	}
	return buf
}

func (j *JsonConfig) Pretty() string {
	return string(j.Json())
}
