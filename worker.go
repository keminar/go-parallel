package parallel

import (
	"sync"
)

//worker map
type workerMap struct {
	sync.RWMutex
	data map[int]struct{}
}

func newWorkerMap() *workerMap {
	m := new(workerMap)
	m.data = make(map[int]struct{})
	return m
}

func (m *workerMap) length() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.data)
}
