package status

import (
	"encoding/json"
	"github.com/plexsysio/taskmanager"
	"go.uber.org/fx"
	"net/http"
	"sync"
)

var Module = fx.Provide(New)

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
	mux *http.ServeMux,
) Manager {
	m := &impl{
		tm: tm,
	}
	mux.HandleFunc("/v1/status", m.httpHandler)
	return m
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

func (m *impl) httpHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := json.MarshalIndent(m.Status(), "", "\t")
	if err != nil {
		http.Error(w, "Failed to get status Err:"+err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(buf)
}
