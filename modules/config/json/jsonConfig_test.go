package jsonConf

import (
	"testing"
)

func getString(val interface{}) string {
	if val == nil {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

func getInt32(val interface{}) int32 {
	if val == nil {
		return -1
	}
	floatV, ok := val.(float64)
	if !ok {
		return -1
	}
	return int32(floatV)
}

func TestGetSetConfig(t *testing.T) {
	conf := &JsonConfig{
		DeviceName: "testDevice",
		Storage:    5,
		MaxPeers:   100,
	}

	verifyFn := func(name string, store, peers int32) {
		t.Logf("Verifying %s %d %d", name, store, peers)
		dName := getString(conf.Get("device_name"))
		if dName != name {
			t.Fatalf("Got incorrect device %s", dName)
		}

		storage := getInt32(conf.Get("storage"))
		if storage != store {
			t.Fatalf("Got incorrect storage %d", storage)
		}

		maxPeers := getInt32(conf.Get("max_peers"))
		if maxPeers != peers {
			t.Fatalf("Got incorrect storage %d", storage)
		}
	}

	verifyFn("testDevice", 5, 100)

	conf.Set("device_name", "newTestDevice")
	conf.Set("storage", int32(10))
	conf.Set("max_peers", int32(200))

	verifyFn("newTestDevice", 10, 200)
}
