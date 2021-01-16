package jsonConf

import (
	"os"
	"testing"

	logger "github.com/ipfs/go-log/v2"
)

var logg = logger.Logger("jsonConfig")

func TestMain(m *testing.M) {
	logger.SetLogLevel("*", "Debug")
	code := m.Run()
	os.Exit(code)
}

type Storage struct {
	Allocated int
}

type Location struct {
	City string
}

func TestGetSetConfig(t *testing.T) {
	var conf = &JsonConfig{}

	var deviceName = "newTestDevice"
	conf.Set("device_name", deviceName)

	var bootstraps = []string{"hello", "JSONtest"}
	conf.Set("bootstraps", bootstraps)

	var i64 = int64(123)
	conf.Set("storage64", i64)
	var f32 = float32(0.101)
	conf.Set("storage32", f32)

	loc := Location{City: "Kolkata"}
	conf.Set("location", loc)

	st := Storage{Allocated: 123}
	conf.Set("storage", st)

	result, _ := conf.Marshal()
	conf.Unmarshal(result)

	var temp string
	conf.Get("device_name", &temp)
	logg.Debug("Got device name : ", temp)
	if temp != "newTestDevice" {
		t.Fatal("Got incorrect device")
	}

	var tempArr []string
	conf.Get("bootstraps", &tempArr)
	logg.Debug("Got bootstraps : ", tempArr)
	if equal(bootstraps, tempArr) != true {
		t.Fatal("Got incorrect bootstrap")
	}

	var tempf64 int64
	conf.Get("storage64", &tempf64)
	logg.Debug("Got storage64 name : ", tempf64)
	if tempf64 != 123 {
		t.Fatal("Got incorrect storage64")
	}

	var temp32 float64
	conf.Get("storage32", &temp32)
	logg.Debug("Got storage32 name : ", temp32)
	if temp32 != 0.101 {
		t.Fatal("Got incorrect storage32")
	}

	locGet := Location{}
	conf.Get("location", &locGet)
	logg.Debug("Got location : ", locGet.City)
	if locGet.City != loc.City {
		t.Fatal("Got incorrect location")
	}

	stGet := Storage{}
	conf.Get("storage", &stGet)
	logg.Debug("Got storage : ", stGet.Allocated)
	if stGet.Allocated != st.Allocated {
		t.Fatal("Got incorrect storage")
	}

	var asdasd interface{}
	conf.Get("device_name", &asdasd)
	logg.Debug("Got storage : ", asdasd)
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		logg.Debug(len(a), len(b))
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
