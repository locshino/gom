// main.go
// Entry point for the gom CLI application.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec" // Needed for Go check
	"strings" // Needed for error checking
	"time"    // Needed for execution duration measurement
)

// AppVersion is defined in config.go

// runGoVersion executes `go version` and returns the output string or an error.
// This is the core check function, separated from printing/caching logic.
func runGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("the 'go' command was not found in your system's PATH. Please install Go before continuing: https://go.dev/doc/install")
		}
		return "", fmt.Errorf("error running 'go version': %w", err)
	}
	return string(output), nil
}

// checkAndCacheGoVersion checks for Go installation, using and updating the global cache.
// Returns the Go version string found (from cache or execution) and any fatal error.
func checkAndCacheGoVersion() (string, bool, error) {
	cachedVersion, found := getGlobalCachedGoVersion() // From cache.go
	if found {
		fmt.Printf("Info: Using cached Go version: %s", cachedVersion) // Includes newline from output
		return cachedVersion, false, nil                              // Found in cache, cache not modified
	}

	// Not in cache, run the command
	fmt.Println("Info: Go version not cached, running `go version`...")
	versionOutput, err := runGoVersion()
	if err != nil {
		return "", false, err // Go not found or command failed
	}

	// Successfully ran, update cache
	fmt.Printf("Info: Found Go version: %s", versionOutput) // Includes newline
	err = updateGlobalCachedGoVersion(versionOutput)        // Update in-memory cache
	if err != nil {
		// Log warning but proceed, caching is not critical failure
		fmt.Fprintf(os.Stderr, "Warning: Failed to update Go version in cache: %v\n", err)
	}

	return versionOutput, true, nil // Found via execution, cache modified
}

// setupProjectContextIfNeeded finds project root and initializes gom.json if needed.
func setupProjectContextIfNeeded(command string) (projectRoot string, setupErr error) {
	needsContext := false
	switch command {
	case "install", "i", "get", "show", "uninstall", "env": // Added 'env'
		needsContext = true
	}
	if !needsContext { return "", nil }

	projectRoot, setupErr = findProjectRoot()
	if setupErr != nil { return "", setupErr }

	setupErr = initProjectGomFile(projectRoot)
	if setupErr != nil {
		// Allow continuing if file not found (except for invalid JSON).
		if !strings.Contains(setupErr.Error(), "not found") && command != "show" && command != "env" { // Allow 'env list' even if file invalid? Maybe not.
			return projectRoot, fmt.Errorf("error loading project state: %v\n Please fix or remove '%s'", setupErr, gomFileName)
		}
		setupErr = nil // Clear non-fatal "not found" error.
	}
	return projectRoot, setupErr
}


func main() {
	startTime := time.Now()

	// --- Handle version flag/command early ---
	if len(os.Args) == 2 && (os.Args[1] == "version" || os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("gom version %s\n", AppVersion)
		os.Exit(0)
	}

	// 1. Check Go installation using cache.
	fmt.Println("Checking Go installation...")
	_, modifiedGoVersionCache, err := checkAndCacheGoVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// We have the goVersion string if needed later, and know if cache was updated.

	// 2. Initialize/Load the GLOBAL alias cache.
	initCache() // Loads aliases, Go version was handled above.

	// --- Command Line Argument Parsing ---
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	command := os.Args[1]
	args := os.Args[2:]

	// 3. Setup Project Context if needed.
	projectRoot, setupErr := setupProjectContextIfNeeded(command)
	if setupErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", setupErr)
		os.Exit(1)
	}

	// --- Command Dispatch ---
	modifiedGlobalAliasCache, modifiedProjectGom, commandErr := dispatchCommand(command, args, projectRoot)

	// --- Error Handling & File Saving ---
	if commandErr != nil {
		if commandErr.Error() != "unrecognized command" {
			fmt.Fprintf(os.Stderr, "Error: %v\n", commandErr)
		}
		os.Exit(1)
	}

	// Save the GLOBAL cache if aliases OR Go version were modified.
	// Don't save after 'cache clear'.
	if (modifiedGlobalAliasCache || modifiedGoVersionCache) && command != "cache" {
		err := saveCache()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save global cache file: %v\n", err)
		}
	}

	// Save the project's gom.json file if it was modified.
	if modifiedProjectGom {
		err := saveProjectGomFile()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save project '%s' file: %v\n", gomFileName, err)
		}
	}

	// Print execution duration.
	fmt.Printf("\nFinished in %s\n", time.Since(startTime).Round(time.Millisecond))
}
