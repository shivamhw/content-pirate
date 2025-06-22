package kv

type KV interface {
	Get(ns string, key string)([]byte, error)
	Set(ns string, key string, val []byte) error
	Del(ns string, key string) error
}