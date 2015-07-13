package main

import (
	"sync"
	. "types"
)

var (
	_session_manager session_manager
)

func init() {
	_session_manager.lists = make(map[int32]*Session)
}

type session_manager struct {
	lists map[int32]*Session
	sync.Mutex
}

func (m *session_manager) get(uid int32) *Session {
	m.Lock()
	defer m.Unlock()
	return m.lists[uid]
}

func (m *session_manager) register(uid int32, sess *Session) {
	m.Lock()
	defer m.Unlock()
	m.lists[uid] = sess
}
func (m *session_manager) unregister(uid int32) {
	m.Lock()
	defer m.Unlock()
	delete(m.lists, uid)
}

//TODO persistence session to mongodb here.
