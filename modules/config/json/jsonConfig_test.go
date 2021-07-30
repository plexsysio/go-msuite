package jsonConf_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"

	"github.com/plexsysio/go-msuite/modules/config/json"
	"github.com/plexsysio/go-msuite/utils"
)

func TestConfig(t *testing.T) {

	t.Run("set then get", func(t *testing.T) {
		conf := jsonConf.DefaultConfig()

		type structVal struct {
			Name string
			Val  int
		}
		conf.Set("Name", "dummy")
		conf.Set("Int", 10)
		conf.Set("StrArray", []string{"hello", "world"})
		conf.Set("Int64", int64(100))
		conf.Set("Float64", float64(1.0000))
		conf.Set("Struct", structVal{
			Name: "struct",
			Val:  100,
		})

		assert := func(val bool, msg string) {
			if !val {
				t.Fatal(msg)
			}
		}

		var strVal string
		assert(conf.Get("Name", &strVal), "getting strval")
		if strVal != "dummy" {
			t.Fatal("invalid value expected: dummy found:", strVal)
		}

		var intVal int
		assert(conf.Get("Int", &intVal), "getting intval")
		if intVal != 10 {
			t.Fatal("invalid value expected: 10 found:", intVal)
		}

		var strArray []string
		assert(conf.Get("StrArray", &strArray), "getting strarray")
		if !equal(strArray, []string{"hello", "world"}) {
			t.Fatal("invalid array found:", strArray)
		}

		var lIntVal int64
		assert(conf.Get("Int64", &lIntVal), "getting int64 val")
		if lIntVal != int64(100) {
			t.Fatal("invalid int64 expected: 100 found:", lIntVal)
		}

		var floatVal float64
		assert(conf.Get("Float64", &floatVal), "getting float val")
		if floatVal != float64(1.0000) {
			t.Fatal("invalid float64 expected: 1.0000 found:", floatVal)
		}

		strct := structVal{}
		assert(conf.Get("Struct", &strct), "getting struct val")
		if strct.Name != "struct" || strct.Val != 100 {
			t.Fatal("invalid struct read", strct)
		}

		// Key not present
		var strVal2 string
		assert(!conf.Get("NonKey", &strVal2), "key should not exist")

		assert(!conf.Get("Int", &strVal2), "incorrect type for Get should fail")
	})

	t.Run("is set", func(t *testing.T) {
		conf := jsonConf.DefaultConfig()
		conf.Set("Bool", true)

		if !conf.IsSet("Bool") {
			t.Fatal("expected Bool to be set")
		}

		if conf.IsSet("NotBool") {
			t.Fatal("expected NotBool to be not set")
		}
	})

	t.Run("exists", func(t *testing.T) {
		conf := jsonConf.DefaultConfig()
		conf.Set("StrVal", "str")

		if !conf.Exists("StrVal") {
			t.Fatal("expected StrVal to exist")
		}

		if conf.Exists("IntVal") {
			t.Fatal("expected IntVal to not exist")
		}
	})

	t.Run("reader writer", func(t *testing.T) {
		conf := jsonConf.DefaultConfig()

		// Create a larger config for Read Write test as they should do multiple
		// reads/writes
		for i := 0; i < 100; i++ {
			randBytes := make([]byte, 32)
			rand.Read(randBytes)
			conf.Set(fmt.Sprintf("key_%d", i+1), randBytes)
		}

		rdr, err := conf.Reader()
		if err != nil {
			t.Fatal(err)
		}

		newConf := &jsonConf.JsonConfig{}
		writer := newConf.Writer()

		_, err = io.Copy(writer, rdr)
		if err != nil {
			t.Fatal(err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatal(err)
		}

		if conf.String() != newConf.String() {
			t.Fatal("string values dont match expected:", conf.String(), "found:", newConf.String())
		}

		if conf.Pretty() != newConf.Pretty() {
			t.Fatal("pretty string values dont match expected:", conf.Pretty(), "found:", newConf.Pretty())
		}

		for i := 0; i < 100; i++ {
			var buf1 []byte
			var buf2 []byte

			key := fmt.Sprintf("key_%d", i+1)

			conf.Get(key, &buf1)
			newConf.Get(key, &buf2)

			if bytes.Compare(buf1, buf2) != 0 {
				t.Fatal("values mismatch for key:", key, "expected:", buf1, "found:", buf2)
			}
		}
	})

	t.Run("read write with file", func(t *testing.T) {
		conf := jsonConf.DefaultConfig()

		// Create a larger config for Read Write test as they should do multiple
		// reads/writes
		for i := 0; i < 10; i++ {
			randBytes := make([]byte, 32)
			rand.Read(randBytes)
			conf.Set(fmt.Sprintf("key_%d", i+1), randBytes)
		}

		rdr, err := conf.Reader()
		if err != nil {
			t.Fatal(err)
		}

		err = utils.WriteToFile(rdr, "newfile.json")
		if err != nil {
			t.Fatal(err)
		}

		defer os.RemoveAll("newfile.json")

		newConf, err := jsonConf.FromFile("newfile.json")
		if err != nil {
			t.Fatal(err)
		}

		if conf.String() != newConf.String() {
			t.Fatal("string values dont match expected:", conf.String(), "found:", newConf.String())
		}

		if conf.Pretty() != newConf.Pretty() {
			t.Fatal("pretty string values dont match expected:", conf.Pretty(), "found:", newConf.Pretty())
		}

		for i := 0; i < 10; i++ {
			var buf1 []byte
			var buf2 []byte

			key := fmt.Sprintf("key_%d", i+1)

			conf.Get(key, &buf1)
			newConf.Get(key, &buf2)

			if bytes.Compare(buf1, buf2) != 0 {
				t.Fatal("values mismatch for key:", key, "expected:", buf1, "found:", buf2)
			}
		}
	})

	t.Run("incompatible value", func(t *testing.T) {
		conf := jsonConf.DefaultConfig()
		conf.Set("Function", func() { fmt.Println("hello") })

		var funcType func()
		found := conf.Get("Function", &funcType)
		if found {
			t.Fatal("non-encodable type should not be found")
		}
	})
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
