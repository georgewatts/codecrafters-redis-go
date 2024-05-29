package main

import (
	"fmt"
	"time"
)

type storeValue struct {
	value string
	ttl   int64
}

type StringStoreService struct {
	store map[string]storeValue
}

type Store interface {
	Set(string, string, int64)
	Get(string) string
}

func (storeService *StringStoreService) Get(key string) string {
	entry := storeService.store[key]

	if entry.ttl > 0 && entry.ttl < time.Now().UnixMilli() {
		delete(storeService.store, key)
		return ""
	}
	fmt.Printf("GET entry: %v\n", entry)
	return entry.value
}

func (storeService *StringStoreService) Set(key string, val string, ttl int64) {
	newEntry := storeValue{
		value: val,
	}

	if ttl > 0 {
		expiry := time.Now().UnixMilli() + ttl
		newEntry.ttl = expiry
	}

	fmt.Printf("SET newEntry: %v\n", newEntry)
	storeService.store[key] = newEntry
}

func NewStringStoreService() *StringStoreService {
	return &StringStoreService{
		store: map[string]storeValue{},
	}
}
