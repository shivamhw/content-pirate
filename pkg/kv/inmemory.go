package kv

import (
	"fmt"
	"log/slog"
	"sync"
)

type bkt map[string][]byte

type InMemDb struct {
	db map[string]bkt
	l *sync.Mutex
}

func GetInMemoryKv() (*InMemDb) {
	return &InMemDb{
		db: make(map[string]bkt),
		l: &sync.Mutex{},
	}
}

func (i *InMemDb) Get(ns string, key string) ([]byte, error) {
	defer i.l.Unlock()
	i.l.Lock()
	var b bkt
	var ok bool
	if b, ok = i.db[ns]; !ok {
		// return nil, fmt.Errorf("namespace not found %s", ns)
		slog.Warn("ns not found creating one, ", "ns", ns)
		i.db[ns] = make(bkt)
		b = i.db[ns]
	}
	if v, ok := b[key]; !ok {
		return nil, fmt.Errorf("key not found, ns %s key %s", ns, key)
	} else {
		return v, nil
	}
}

func (i *InMemDb) Set(ns string, key string, val []byte) (error) {
	defer i.l.Unlock()
	i.l.Lock()
	var b bkt
	var ok bool
	if b, ok = i.db[ns]; !ok {
		slog.Warn("ns not found creating one, ", "ns", ns)
		i.db[ns] = make(bkt)
		b = i.db[ns]
	}
	b[key] = val
	return nil
}

func (i *InMemDb) Del(ns string, key string) (error) {
	var b bkt
	var ok bool
	if b, ok = i.db[ns]; !ok {
		return fmt.Errorf("namespace not found %s", ns)
	}
	delete(b, key)
	return nil
}
