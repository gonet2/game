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
	defer r.Unlock()
	r.records[id] = v
}

// unregister a user
func (r *Registry) Unregister(id int32) {
	r.Lock()
	defer r.Unlock()
	delete(r.records, id)
}

// query a user
func (r *Registry) Query(id int32) interface{} {
	r.RLock()
	defer r.RUnlock()
	return r.records[id]
}

// return count of online users
func (r *Registry) Count() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.records)
}

func Register(id int32, v interface{}) {
	_default_registry.Register(id, v)
}

func Unregister(id int32) {
	_default_registry.Unregister(id)
}

func Query(id int32) interface{} {
	return _default_registry.Query(id)
}

func Count() int {
	return _default_registry.Count()
}
