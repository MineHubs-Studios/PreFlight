package utils

import (
	"os"
	"sync"
)

// fileExistenceCache provides a simple cache for file existence checks.
var fileExistenceCache = struct {
	sync.RWMutex
	cache map[string]bool
}{cache: make(map[string]bool)}

// FileExists checks if a file exists and caches results.
func FileExists(path string) bool {
	// Check the cache first.
	fileExistenceCache.RLock()
	exists, found := fileExistenceCache.cache[path]
	fileExistenceCache.RUnlock()

	if found {
		return exists
	}

	// If cache is missing then check the filesystem.
	_, err := os.Stat(path)
	exists = err == nil

	// Update cache.
	fileExistenceCache.Lock()
	fileExistenceCache.cache[path] = exists
	fileExistenceCache.Unlock()

	return exists
}
