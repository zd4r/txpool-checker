package txpool

import "sync"

type Txpool struct {
	mx sync.RWMutex
	m  map[string]string
}

func New() *Txpool {
	return &Txpool{
		m: make(map[string]string),
	}
}

func (tp *Txpool) Load(key string) (string, bool) {
	tp.mx.RLock()
	defer tp.mx.RUnlock()
	val, ok := tp.m[key]
	return val, ok
}

func (tp *Txpool) Store(key, value string) {
	tp.mx.Lock()
	defer tp.mx.Unlock()
	tp.m[key] = value
}

func (tp *Txpool) Delete(key string) {
	tp.mx.Lock()
	defer tp.mx.Unlock()
	delete(tp.m, key)
}

func (tp *Txpool) Count() int {
	tp.mx.Lock()
	defer tp.mx.Unlock()
	return len(tp.m)
}
