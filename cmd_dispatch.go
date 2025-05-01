// cmd_dispatch.go
// Handles routing commands to their specific implementation functions (handlers).
package main

import (
	"fmt"
)

// dispatchCommand selects and executes the appropriate command handler.
// Returns flags indicating if global cache or project gom.json were potentially modified, and any error.
func dispatchCommand(command string, args []string, projectRoot string) (modifiedGlobalCache bool, modifiedProjectGom bool, commandErr error) {

	// Check if the command requires being run inside a Go project.
	needsProjectContext := false
	switch command {
	case "install", "i", "get", "show", "uninstall", "env": // Added 'env'
		needsProjectContext = true
	}

	if needsProjectContext && projectRoot == "" && command != "env" { // Allow 'env' maybe? No, env modifes gom.json.
        // If context is needed but not available (projectRoot is empty), return error early.
		// Exception: Maybe allow 'env list' globally? For now, require project context.
		commandErr = fmt.Errorf("command '%s' must be run inside a Go module (go.mod not found)", command)
		return
	}


	// Dispatch based on command name.
	switch command {
	case "init":
		commandErr = handleInit(args)
		if commandErr == nil { modifiedProjectGom = true }

	case "install", "i":
		var modified bool
		modified, commandErr = handleInstall(args)
		if commandErr == nil && modified { modifiedProjectGom = true }

	case "get":
		modifiedGlobalCache, modifiedProjectGom, commandErr = handleGetAs(args)

	case "show":
		commandErr = handleShow(args)

	case "uninstall":
		var modified bool
		modified, commandErr = handleUninstall(args)
		if commandErr == nil && modified { modifiedProjectGom = true }

	case "env": // New command group for environments
		var modified bool
		modified, commandErr = handleEnvCommand(args) // Dispatch to env subcommand handler.
		if commandErr == nil && modified { modifiedProjectGom = true }

	case "doctor", "check": // New command group for checks
		var modifiedCache bool
		modifiedCache, commandErr = handleDoctorCommand(args) // Dispatch to doctor subcommand handler.
		if commandErr == nil && modifiedCache { modifiedGlobalCache = true }


	case "cache":
		// Cache commands don't modify project gom.json or global cache directly
		// in a way that needs tracking *here* (clear handles its own saving logic).
		commandErr = handleCacheCommand(args)

	// case "version", "-v", "--version": // Handled in main.
	default:
		fmt.Printf("Error: Unrecognized command '%s'.\n\n", command)
		printUsage() // Defined in utils.go
		commandErr = fmt.Errorf("unrecognized command") // Signal error to main.
	}

	return // Return collected flags and error.
}

// handleCacheCommand dispatches cache subcommands (list, clear).
func handleCacheCommand(args []string) error {
	if len(args) == 0 { return fmt.Errorf("cache command requires a subcommand (list or clear)") }
	subCommand := args[0]
	subArgs := args[1:]
	switch subCommand {
	case "list": return handleCacheList(subArgs)
	case "clear": return handleCacheClear(subArgs)
	default: return fmt.Errorf("unrecognized cache subcommand '%s'. Available: list, clear", subCommand)
	}
}

// handleEnvCommand dispatches environment subcommands (add, remove, list).
// Returns true if gom.json was modified, and any error.
func handleEnvCommand(args []string) (bool, error) {
	if projectGomFile == nil || projectGomFile.filePath == "" {
		// This check should be redundant due to main, but good practice.
        return false, fmt.Errorf("env commands must be run inside a Go module")
    }
	if len(args) == 0 { return false, fmt.Errorf("env command requires a subcommand (add, remove, list)") }

	subCommand := args[0]
	subArgs := args[1:]
	modified := false
	var err error

	switch subCommand {
	case "add":
		modified, err = handleEnvAdd(subArgs)
	case "remove", "rm":
		modified, err = handleEnvRemove(subArgs)
	case "list", "ls":
		err = handleEnvList(subArgs)
	default:
		err = fmt.Errorf("unrecognized env subcommand '%s'. Available: add, remove, list", subCommand)
	}
	return modified, err
}

// handleDoctorCommand dispatches check/doctor subcommands.
// Returns true if the global cache was modified, and any error.
func handleDoctorCommand(args []string) (bool, error) {
	if len(args) == 0 {
		// Default action: check Go version
		return handleCheckGo(args)
	}
	subCommand := args[0]
	subArgs := args[1:]
	switch subCommand {
	case "go":
		return handleCheckGo(subArgs)
	// Add other checks here later if needed
	default:
		return false, fmt.Errorf("unrecognized check/doctor subcommand '%s'. Available: go", subCommand)
	}
}
