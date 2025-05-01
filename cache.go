// cache.go
// Manages the global alias and Go version cache.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// globalCacheFileName is defined in config.go

// GlobalCacheData struct holds the globally cached information.
type GlobalCacheData struct {
	// Aliases maps alias@version -> import_path for globally shared aliases.
	Aliases map[string]string `json:"aliases,omitempty"` // Use omitempty if map can be nil/empty
	// GoVersion stores the string output of `go version`.
	GoVersion string `json:"go_version,omitempty"`
	mu        sync.RWMutex // Protects concurrent access.
}

// globalCache holds the loaded global cache data in memory.
var globalCache *GlobalCacheData

// globalCacheDirPath is defined in config.go

// getGlobalCacheFilePath returns the absolute path to the global cache file.
func getGlobalCacheFilePath() (string, error) {
	dirPath, err := cacheDirPath() // Use function from config.go
	if err != nil {
		return "", err
	}
	return filepath.Join(dirPath, globalCacheFileName), nil // Use constant from config.go
}


// initCache loads the global cache from its file on program startup.
func initCache() {
	globalCache = &GlobalCacheData{
		Aliases: make(map[string]string), // Initialize map
	}

	cacheFilePath, err := getGlobalCacheFilePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not determine global cache file path: %v\n", err)
		return
	}

	// Read the cache file.
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Warning: Could not read global cache file '%s': %v\n", cacheFilePath, err)
		}
		return // Use empty cache if not found or error reading.
	}

	// Parse JSON data.
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	if err := json.Unmarshal(data, &globalCache); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not parse global cache file '%s', starting with empty cache. Error: %v\n", cacheFilePath, err)
		// Reset on parse error, keeping the initialized map.
		globalCache.Aliases = make(map[string]string)
		globalCache.GoVersion = ""
		return
	}

	// Ensure map is initialized after unmarshaling.
	if globalCache.Aliases == nil {
		globalCache.Aliases = make(map[string]string)
	}
	// Note: GoVersion will be its zero value (empty string) if not present in JSON.
}

// saveCache writes the current in-memory global cache data to the file.
func saveCache() error {
	if globalCache == nil {
		return fmt.Errorf("internal error: global cache not initialized before saving")
	}
	cacheFilePath, err := getGlobalCacheFilePath()
	if err != nil {
		return fmt.Errorf("could not get global cache file path for saving: %w", err)
	}

	// Marshal the entire struct.
	globalCache.mu.RLock()
	data, err := json.MarshalIndent(globalCache, "", "  ")
	globalCache.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("could not marshal global cache data to JSON: %w", err)
	}

	// Write to file.
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	if err := os.WriteFile(cacheFilePath, data, 0640); err != nil { // Permissions 0640: rw-r-----.
		return fmt.Errorf("could not write to global cache file '%s': %w", cacheFilePath, err)
	}
	return nil
}

// getGlobalCachedPath retrieves the import path for a specific alias@version from the global cache.
func getGlobalCachedPath(aliasVersion string) (string, bool) {
	if globalCache == nil { return "", false }
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
    if globalCache.Aliases == nil { return "", false }
	path, found := globalCache.Aliases[aliasVersion]
	return path, found && path != ""
}

// addGlobalCachedPath adds or updates an alias in the in-memory global cache.
func addGlobalCachedPath(aliasVersion, importPath string) error {
	if globalCache == nil { return fmt.Errorf("internal error: global cache not initialized") }
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	if globalCache.Aliases == nil { globalCache.Aliases = make(map[string]string) }
	globalCache.Aliases[aliasVersion] = importPath
	return nil
}

// getAllGlobalCachedAliases returns a copy of the global cache map.
func getAllGlobalCachedAliases() (map[string]string, error) {
    if globalCache == nil { return nil, fmt.Errorf("internal error: global cache not initialized") }
    globalCache.mu.RLock()
    defer globalCache.mu.RUnlock()
    if globalCache.Aliases == nil { return make(map[string]string), nil }
    aliasesCopy := make(map[string]string, len(globalCache.Aliases))
    for k, v := range globalCache.Aliases { aliasesCopy[k] = v }
    return aliasesCopy, nil
}

// clearGlobalCache resets the in-memory cache and attempts to save the empty cache to disk.
func clearGlobalCache() error {
	if globalCache == nil { return fmt.Errorf("internal error: global cache not initialized") }
	globalCache.mu.Lock()
	globalCache.Aliases = make(map[string]string)
	globalCache.GoVersion = "" // Also clear Go version
	globalCache.mu.Unlock()
	fmt.Println("Info: In-memory global cache cleared.")
	err := saveCache() // Attempt to save the now empty cache file.
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not overwrite global cache file during clear: %v\n", err)
	}
	return nil
}

// --- Go Version Cache Specific Functions ---

// getGlobalCachedGoVersion retrieves the cached Go version string.
func getGlobalCachedGoVersion() (string, bool) {
	if globalCache == nil { return "", false }
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	found := globalCache.GoVersion != ""
	return globalCache.GoVersion, found
}

// updateGlobalCachedGoVersion updates the Go version in the in-memory global cache.
// Does not save the file to disk.
func updateGlobalCachedGoVersion(version string) error {
	if globalCache == nil { return fmt.Errorf("internal error: global cache not initialized") }
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.GoVersion = version
	return nil
}
