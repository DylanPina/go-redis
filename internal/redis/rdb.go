package redis

import (
	"sync"
	"time"
)

type cacheItem struct {
	value      string
	expiration int64 // Expiration time in milliseconds since epoch
}

var (
	cache = make(map[string]cacheItem)
	mu    sync.RWMutex
)

// Set stores the key-value pair with expiration in px milliseconds from now.
func Set(key, value string, px int64) {
	mu.Lock()
	defer mu.Unlock()

	var expiration int64
	if px >= 0 {
		expiration = time.Now().UnixMilli() + px
	} else {
		expiration = -1 // No expiration
	}

	cache[key] = cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Get retrieves the value for a key, only if it hasn't expired.
// Returns (value, true) if present and not expired.
// Returns ("", false) if not found or expired// Get returns (value, true) if key exists and is not expired.
func Get(key string) (string, bool) {
	mu.RLock()
	item, exists := cache[key]
	mu.RUnlock()

	if !exists {
		return "", false
	}

	if item.expiration != -1 && time.Now().UnixMilli() > item.expiration {
		mu.Lock()
		delete(cache, key)
		mu.Unlock()
		return "", false
	}

	return item.value, true
}
