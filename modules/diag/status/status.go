package status

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/plexsysio/taskmanager"
)

type Manager interface {
	Report(string, Status)
	Status() map[string]interface{}
}

type Status interface {
	Delta(Status)
}

type String string

func (s String) Delta(_ Status) {}

type Map map[string]interface{}

func (m Map) Delta(old Status) {
	oldMp, ok := old.(Map)
	if !ok {
		return
	}
	for k, v := range oldMp {
		_, ok := m[k]
		if !ok {
			m[k] = v
		}
	}
}

type impl struct {
	mp sync.Map
	tm *taskmanager.TaskManager
}

func New(
	tm *taskmanager.TaskManager,
) Manager {
	m := &impl{
		tm: tm,
	}
	return m
}

func RegisterHTTP(m Manager, mux *http.ServeMux) {
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
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

func (m *impl) Report(key string, msg Status) {
	val, loaded := m.mp.LoadOrStore(key, msg)
	if !loaded {
		return
	}
	oldStatus, ok := val.(Status)
	if ok {
		msg.Delta(oldStatus)
		m.mp.Store(key, msg)
	}
}

func (m *impl) Status() map[string]interface{} {
	retStatus := make(map[string]interface{})
	m.mp.Range(func(k, v interface{}) bool {
		retStatus[k.(string)] = v
		return true
	})
	retStatus["Task Manager"] = m.tm.TaskStatus()
	return retStatus
}
