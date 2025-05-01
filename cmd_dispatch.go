// cmd_dispatch.go
// Handles routing commands to their specific implementation functions (handlers).
package main

import (
	"fmt"
)

// dispatchCommand selects and executes the appropriate command handler.
//
// It takes the command name, arguments, and the project root path (if applicable).
// It checks if project context (projectRoot) is required and valid for the command.
// It calls the corresponding handle... function for the command.
// Returns flags indicating if global cache or project gom.json were potentially modified,
// and any error returned by the handler.
func dispatchCommand(command string, args []string, projectRoot string) (modifiedGlobalCache bool, modifiedProjectGom bool, commandErr error) {

	// Check if the command requires being run inside a Go project.
	needsProjectContext := false
	switch command {
	case "install", "i", "get", "show", "uninstall":
		needsProjectContext = true
	}

	// If context is needed but not available (projectRoot is empty), return error early.
	if needsProjectContext && projectRoot == "" {
		// This error should ideally be caught in main, but double-checking protects handlers.
		commandErr = fmt.Errorf("command '%s' must be run inside a Go module (go.mod not found)", command)
		return // Return immediately.
	}


	// Dispatch to the appropriate handler based on the command name.
	switch command {
	case "init":
		// Init operates on the current directory.
		commandErr = handleInit(args)
		if commandErr == nil { modifiedProjectGom = true } // Init modifies gom.json.

	case "install", "i":
		// Context checked above.
		var modified bool
		modified, commandErr = handleInstall(args) // Calls handler from commands.go.
		if commandErr == nil && modified { modifiedProjectGom = true } // Mark if gom.json was modified.

	case "get":
		// Context checked above.
		// handleGetAs returns modification flags directly.
		modifiedGlobalCache, modifiedProjectGom, commandErr = handleGetAs(args) // Calls handler from commands.go.

	case "show":
		// Context checked above.
		commandErr = handleShow(args) // Calls handler from commands.go.

	case "uninstall":
		// Context checked above.
		var modified bool
		modified, commandErr = handleUninstall(args) // Calls handler from commands.go.
		if commandErr == nil && modified { modifiedProjectGom = true } // Mark if gom.json was modified.

	case "cache":
		// Cache commands don't need project context.
		commandErr = handleCacheCommand(args) // Dispatch to cache subcommand handler below.

	// case "version", "-v", "--version": // Handled in main.
	default:
		// Unrecognized command.
		fmt.Printf("Error: Unrecognized command '%s'.\n\n", command)
		printUsage() // Defined in utils.go
		commandErr = fmt.Errorf("unrecognized command") // Signal error to main.
	}

	return // Return collected flags and error.
}

// handleCacheCommand dispatches cache subcommands (list, clear).
func handleCacheCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("cache command requires a subcommand (list or clear)")
	}
	subCommand := args[0]
	subArgs := args[1:] // Arguments for the subcommand (currently none).

	switch subCommand {
	case "list":
		return handleCacheList(subArgs) // Calls handler from commands.go.
	case "clear":
		return handleCacheClear(subArgs) // Calls handler from commands.go.
	default:
		return fmt.Errorf("unrecognized cache subcommand '%s'. Available: list, clear", subCommand)
	}
}
