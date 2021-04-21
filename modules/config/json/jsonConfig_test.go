package jsonConf

import (
	"github.com/plexsysio/go-msuite/utils"
	"os"
	"testing"
)

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

	defer func() {
		os.Remove("tmp.json")
	}()

	err := utils.WriteToFile(conf, "tmp.json")
	if err != nil {
		t.Fatal("Failed writing config to file")
	}
	conf2 := &JsonConfig{}
	err = utils.ReadFromFile(conf2, "tmp.json")
	if err != nil {
		t.Fatal("Failed reading config from file")
	}

	var temp string
	conf2.Get("device_name", &temp)
	if temp != "newTestDevice" {
		t.Fatal("Got incorrect device")
	}

	var tempArr []string
	conf2.Get("bootstraps", &tempArr)
	if equal(bootstraps, tempArr) != true {
		t.Fatal("Got incorrect bootstrap")
	}

	var tempf64 int64
	conf2.Get("storage64", &tempf64)
	if tempf64 != 123 {
		t.Fatal("Got incorrect storage64")
	}

	var temp32 float64
	conf2.Get("storage32", &temp32)
	if temp32 != 0.101 {
		t.Fatal("Got incorrect storage32")
	}

	locGet := Location{}
	conf2.Get("location", &locGet)
	if locGet.City != loc.City {
		t.Fatal("Got incorrect location")
	}

	stGet := Storage{}
	conf2.Get("storage", &stGet)
	if stGet.Allocated != st.Allocated {
		t.Fatal("Got incorrect storage")
	}

}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
