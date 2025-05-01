// main.go
// Entry point for the gom CLI application. Handles initial setup,
// argument parsing, command dispatching, and final file saving.
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

// checkGoInstallation verifies if the 'go' command is available in the system's PATH.
//
// It executes `go version` and checks for errors, specifically command not found.
// Prints the found Go version upon success.
// Returns an error if Go is not found or if the command fails.
func checkGoInstallation() error {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("the 'go' command was not found in your system's PATH. Please install Go before continuing: https://go.dev/doc/install")
		}
		return fmt.Errorf("error running 'go version': %w", err)
	}
	fmt.Printf("Info: Found Go version: %s", string(output)) // Output includes newline.
	return nil
}

// setupProjectContextIfNeeded determines if the command needs project context,
// finds the project root (go.mod), and initializes the project's gom.json.
//
// It checks the command name against a list of commands requiring project context.
// If context is needed, it calls findProjectRoot and initProjectGomFile.
// Returns the project root path (or empty string if context not needed/found)
// and any fatal error encountered during setup (e.g., invalid gom.json).
func setupProjectContextIfNeeded(command string) (projectRoot string, setupErr error) {
	needsContext := false
	switch command {
	case "install", "i", "get", "show", "uninstall":
		needsContext = true
	}

	if !needsContext {
		return "", nil // No context needed.
	}

	// Find project root (directory containing go.mod).
	projectRoot, setupErr = findProjectRoot() // From gomfile.go
	if setupErr != nil {
		return "", setupErr // Return error if root not found.
	}

	// Load or initialize project-specific gom.json within the found root.
	setupErr = initProjectGomFile(projectRoot) // From gomfile.go
	if setupErr != nil {
		// Allow continuing only if file just doesn't exist yet,
		// or if the command is 'show' (it handles the error itself).
		// Return a fatal error for invalid JSON etc., preventing further execution.
		if !strings.Contains(setupErr.Error(), "not found") && command != "show" {
			return projectRoot, fmt.Errorf("error loading project state: %v\n Please fix or remove '%s'", setupErr, gomFileName)
		}
		// Otherwise, it's just a non-existent file, which is okay for now.
		setupErr = nil // Clear the "not found" error.
	}

	return projectRoot, setupErr // Return root path and potential fatal error.
}

func main() {
	startTime := time.Now()

	// Handle version flag/command immediately.
	if len(os.Args) == 2 && (os.Args[1] == "version" || os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("gom version %s\n", AppVersion)
		os.Exit(0)
	}

	// 1. Check Go installation.
	fmt.Println("Checking Go installation...")
	if err := checkGoInstallation(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize global cache.
	initCache() // From cache.go

	// 3. Parse command line arguments.
	if len(os.Args) < 2 {
		printUsage() // From utils.go
		os.Exit(1)
	}
	command := os.Args[1]
	args := os.Args[2:]

	// 4. Setup Project Context (find root, load gom.json) if needed.
	projectRoot, setupErr := setupProjectContextIfNeeded(command)
	if setupErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", setupErr)
		os.Exit(1)
	}

	// 5. Dispatch the command to its handler.
	// This function (from cmd_dispatch.go) contains the main command logic switch.
	modifiedGlobalCache, modifiedProjectGom, commandErr := dispatchCommand(command, args, projectRoot)

	// 6. Handle command errors.
	if commandErr != nil {
		// Avoid double printing for "unrecognized command" as dispatchCommand already printed usage.
		if commandErr.Error() != "unrecognized command" {
			fmt.Fprintf(os.Stderr, "Error: %v\n", commandErr)
		}
		os.Exit(1)
	}

	// 7. Save modified files.
	// Save global cache if modified.
	if modifiedGlobalCache { // Check flag set by dispatchCommand.
		err := saveCache() // From cache.go
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save global cache file: %v\n", err)
		}
	}
	// Save project gom.json if modified.
	if modifiedProjectGom { // Check flag set by dispatchCommand.
		err := saveProjectGomFile() // From gomfile.go
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save project '%s' file: %v\n", gomFileName, err)
		}
	}

	// 8. Print execution duration.
	fmt.Printf("\nFinished in %s\n", time.Since(startTime).Round(time.Millisecond))
}
