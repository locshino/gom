// config.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Application constants.
const (
	// AppVersion is the current version of the gom tool.
	// It can be set at build time using ldflags:
	// go build -ldflags="-X main.AppVersion=0.1.0" -o gom .
	AppVersion = "0.1.0"

	// Filenames used by gom.
	gomFileName         = "gom.json"   // Project-local aliases.
	globalCacheFileName = "cache.json" // Global alias cache.

	// Directory names.
	appNameDir = "gom" // Subdirectory name within standard config/cache dirs.
)

// --- Path Helper Functions ---

// configDirPath returns the platform-specific config directory path for the app.
// It ensures the directory exists, creating it if necessary.
// It falls back to ~/.gom if the standard directory cannot be determined.
func configDirPath() (string, error) {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to ~/.gom if standard dir fails.
		fmt.Fprintf(os.Stderr, "Warning: Could not get standard user config directory (%v), attempting fallback to ~/.gom\n", err)
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", fmt.Errorf("could not get standard user config directory or home directory: %w, %w", err, homeErr)
		}
		baseConfigDir = filepath.Join(homeDir, "."+appNameDir) // e.g., ~/.gom
	}
	// Create the application's subdirectory.
	appConfigDir := filepath.Join(baseConfigDir, appNameDir)
	if err := os.MkdirAll(appConfigDir, 0750); err != nil { // Ensure directory exists (permissions 0750: rwxr-x---).
		return "", fmt.Errorf("could not create application config directory '%s': %w", appConfigDir, err)
	}
	return appConfigDir, nil
}

// cacheDirPath returns the platform-specific cache directory path for the app.
// It ensures the directory exists, creating it if necessary.
// It falls back to config_dir/cache if the standard directory cannot be determined.
func cacheDirPath() (string, error) {
	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to config_dir/cache if standard cache dir fails.
		fmt.Fprintf(os.Stderr, "Warning: Could not get standard user cache directory (%v), attempting fallback to config_dir/cache\n", err)
		configDir, configErr := configDirPath() // Reuse config dir logic for fallback base.
		if configErr != nil {
			return "", fmt.Errorf("could not get standard user cache directory or fallback config directory: %w, %w", err, configErr)
		}
		baseCacheDir = filepath.Join(configDir, "cache") // e.g. ~/.gom/cache or ~/.config/gom/cache
	}
	// Create the application's subdirectory.
	appCacheDir := filepath.Join(baseCacheDir, appNameDir)
	if err := os.MkdirAll(appCacheDir, 0750); err != nil { // Ensure directory exists (permissions 0750: rwxr-x---).
		return "", fmt.Errorf("could not create application cache directory '%s': %w", appCacheDir, err)
	}
	return appCacheDir, nil
}
