package jsonConf

import (
	"encoding/json"
	logger "github.com/ipfs/go-log"
	"strconv"
)

var log = logger.Logger("jsonConf")

type JsonConfig struct {
	GrpcPort    int32  `json:"grpc_port"`
	P2PPort     int32  `json:"p2p_port"`
	UseJwt      bool   `json:"use_jwt"`
	UseTracing  bool   `json:"use_tracing"`
	ServiceName string `json:"service_name"`
	TracingHost string `json:"tracing_host"`
}

func DefaultConfig() *JsonConfig {
	log.Info("Returning default config")
	return &JsonConfig{
		P2PPort:     10000,
		UseJwt:      false,
		UseTracing:  true,
		TracingHost: "localhost:16656",
		ServiceName: "fxTest",
	}
}

// Functions for Store and Item interfaces
func (j *JsonConfig) GetNamespace() string {
	return "ssConfig"
}

func (j *JsonConfig) GetId() string {
	return "json"
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

func (j *JsonConfig) jsonMap() (map[string]interface{}, error) {
	buf, err := j.Marshal()
	if err != nil {
		return nil, err
	}
	newMp := make(map[string]interface{})
	err = json.Unmarshal(buf, &newMp)
	if err != nil {
		return nil, err
	}
	return newMp, nil
}

func (j *JsonConfig) Get(key string) interface{} {
	jMap, err := j.jsonMap()
	if err != nil {
		return nil
	}
	val, ok := jMap[key]
	if !ok {
		return nil
	}
	if key == "grpc_port" || key == "p2p_port" {
		return int32(val.(float64))
	}
	if key == "use_jwt" || key == "use_tracing" {
		return val.(bool)
	}
	return val.(string)
}

func (j *JsonConfig) Set(key string, val interface{}) {
	switch key {
	case "grpc_port":
		if intVal, ok := val.(int32); ok {
			j.GrpcPort = intVal
		} else if strVal, ok := val.(string); ok {
			iVal, err := strconv.ParseInt(strVal, 10, 0)
			if err == nil {
				j.GrpcPort = int32(iVal)
			}
		}
	case "p2p_port":
		if intVal, ok := val.(int32); ok {
			j.P2PPort = intVal
		} else if strVal, ok := val.(string); ok {
			iVal, err := strconv.ParseInt(strVal, 10, 0)
			if err == nil {
				j.P2PPort = int32(iVal)
			}
		}
	case "use_jwt":
		if boolVal, ok := val.(bool); ok {
			j.UseJwt = boolVal
		}
	case "use_tracing":
		if boolVal, ok := val.(bool); ok {
			j.UseTracing = boolVal
		}
	case "service_name":
		if strVal, ok := val.(string); ok {
			j.ServiceName = strVal
		}
	case "tracing_host":
		if strVal, ok := val.(string); ok {
			j.TracingHost = strVal
		}
	}
}
