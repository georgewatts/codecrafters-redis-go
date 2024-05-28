package main

import (
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
	Set(string, string, int64) string
	Get(string) string
}

func (storeService *StringStoreService) Get(key string) string {
	entry := storeService.store[key]

	if entry.value == "" {
		return NewBulkString("")
	} else if entry.ttl > 0 && entry.ttl < time.Now().UnixMilli() {
		delete(storeService.store, key)
		return NewBulkString("")
	}

	return NewBulkString(entry.value)
}

func (storeService *StringStoreService) Set(key string, val string, ttl int64) string {
	newEntry := storeValue{
		value: val,
	}

	if ttl > 0 {
		expiry := time.Now().UnixMilli() + ttl
		newEntry.ttl = expiry
	}
	storeService.store[key] = newEntry
	return NewBulkString(OK)
}

func NewStringStoreService() *StringStoreService {
	return &StringStoreService{
		store: map[string]storeValue{},
	}
}
