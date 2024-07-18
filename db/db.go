package db

import "sync"

type DB struct {
	mu   sync.Mutex
	data map[string]string
}

func NewDB() *DB {
	// Implementation
	return &DB{}
}

func (db *DB) Get(key string) (string, bool) {
	// Implementation
	return "", false
}

func (db *DB) Set(key, value string) {
	// Implementation
}

func (db *DB) Delete(key string) {
	// Implementation
}
