package status

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Reporter interface {
	Status() interface{}
}

type Manager interface {
	AddReporter(string, Reporter)
	Status() map[string]interface{}
}

type impl struct {
	mp sync.Map
}

func New() Manager {
	return new(impl)
}

func (m *impl) AddReporter(key string, reporter Reporter) {
	m.mp.Store(key, reporter)
}

func (m *impl) Status() map[string]interface{} {
	retStatus := make(map[string]interface{})
	m.mp.Range(func(k, v interface{}) bool {
		retStatus[k.(string)] = v.(Reporter).Status()
		return true
	})
	return retStatus
}

func RegisterHTTP(m Manager, mux *http.ServeMux) {
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		buf, err := json.MarshalIndent(m.Status(), "", "\t")
		if err != nil {
			http.Error(w, "Failed to get status Err:"+err.Error(),
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(buf)
	})
}
