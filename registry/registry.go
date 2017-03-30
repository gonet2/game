package registry

import (
	"sync"
)

type Registry struct {
	records map[int32]interface{} // id -> v
	sync.RWMutex
}

var (
	_default_registry Registry
)

func init() {
	_default_registry.init()
}

func (r *Registry) init() {
	r.records = make(map[int32]interface{})
}

// register a user
func (r *Registry) Register(id int32, v interface{}) {
	r.Lock()
	r.records[id] = v
	r.Unlock()
}

// unregister a user
func (r *Registry) Unregister(id int32, v interface{}) {
	r.Lock()
	if oldv, ok := r.records[id]; ok {
		if oldv == v {
			delete(r.records, id)
		}
	}
	r.Unlock()
}

// query a user
func (r *Registry) Query(id int32) (x interface{}) {
	r.RLock()
	x = r.records[id]
	r.RUnlock()
	return
}

// return count of online users
func (r *Registry) Count() (count int) {
	r.RLock()
	count = len(r.records)
	r.RUnlock()
	return
}

func Register(id int32, v interface{}) {
	_default_registry.Register(id, v)
}

func Unregister(id int32, v interface{}) {
	_default_registry.Unregister(id, v)
}

func Query(id int32) interface{} {
	return _default_registry.Query(id)
}

func Count() int {
	return _default_registry.Count()
}
