// commands.go
// Contains handler functions for each distinct gom command.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// --- Init Handler ---

// handleInit creates an empty gom.json file in the current directory.
func handleInit(args []string) error {
	if len(args) > 0 { return fmt.Errorf("the 'init' command does not take any arguments") }
	currentDir, err := os.Getwd()
	if err != nil { return fmt.Errorf("could not get current working directory: %w", err) }
	gomFilePath := filepath.Join(currentDir, gomFileName)
	if _, err := os.Stat(gomFilePath); err == nil {
		fmt.Printf("'%s' already exists in this directory.\n", gomFileName)
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error checking for existing '%s': %w", gomFileName, err)
	}
	initialGomFile := &GomFile{
		Dependencies: make(map[string]string),
		Environments: make(map[string]EnvironmentInfo), // Also init environments
		filePath:     gomFilePath,
	}
	projectGomFile = initialGomFile // Allow saving
	if err = saveProjectGomFile(); err != nil {
		return fmt.Errorf("failed to create '%s': %w", gomFileName, err)
	}
	fmt.Printf("Initialized empty '%s' in %s\n", gomFileName, currentDir)
	return nil
}

// --- Dependency Alias Handlers ---

// handleInstall installs dependencies based on aliases.
// Returns (modifiedProjectGom bool, error).
func handleInstall(args []string) (bool, error) {
    modifiedProjectGom := false
    if projectGomFile == nil || projectGomFile.filePath == "" {
         return modifiedProjectGom, fmt.Errorf("could not find '%s' or project root (go.mod). Run 'gom init' first?", gomFileName)
    }
	if len(args) == 1 { // Install specific alias
		aliasVersion := args[0]
		packageName, version := parsePackageArg(aliasVersion)
		if packageName == "" || version == "" {
			 return modifiedProjectGom, fmt.Errorf("invalid alias format: '%s'. Must be <name>@<version>", aliasVersion)
		}
		importPath, found := getProjectDependency(aliasVersion)
		if !found { // Check global cache
			fmt.Printf("-> Alias '%s' not found in project '%s', checking global cache...\n", aliasVersion, gomFileName)
			importPath, found = getGlobalCachedPath(aliasVersion)
			if !found {
				return modifiedProjectGom, fmt.Errorf("alias '%s' not found in project '%s' or global cache. Use 'gom get <link> as %s [-g]' to define it first", aliasVersion, gomFileName, aliasVersion)
			}
			// Found globally, add locally
			fmt.Printf("-> Found '%s' in global cache: %s\n", aliasVersion, importPath)
			if err := addProjectDependency(aliasVersion, importPath); err != nil {
				return modifiedProjectGom, fmt.Errorf("internal error adding dependency from global cache to project '%s': %w", gomFileName, err)
			}
			modifiedProjectGom = true
			fmt.Printf("-> Added '%s' -> '%s' to project '%s'.\n", aliasVersion, importPath, gomFileName)
		}
		// Install using resolved path
		fmt.Printf("-> Installing '%s' using path: %s\n", aliasVersion, importPath)
		fmt.Printf("-> Preparing to run: go get %s\n", importPath)
		if err := runGoCommand("get", importPath); err != nil {
			return modifiedProjectGom, fmt.Errorf("'go get %s' command failed", importPath)
		}
		fmt.Printf("\n-> Success! Installed/updated: %s\n", importPath)
	} else if len(args) == 0 { // Install all from project gom.json
		fmt.Printf("-> Installing all dependencies listed in project '%s'...\n", gomFileName)
		dependencies, err := getAllProjectDependencies()
        if err != nil { return modifiedProjectGom, err }
		if len(dependencies) == 0 {
			fmt.Printf("No dependencies found in '%s'. Nothing to install.\n", gomFileName)
			return modifiedProjectGom, nil
		}
		installErrors := []string{}
		successCount := 0
		for aliasVersion, importPath := range dependencies {
			fmt.Printf("--> Installing %s (%s)\n", aliasVersion, importPath)
			if err := runGoCommand("get", importPath); err != nil {
				errorMsg := fmt.Sprintf("Failed to install %s (%s): %v", aliasVersion, importPath, err)
				fmt.Fprintf(os.Stderr, "Error: %s\n", errorMsg)
				installErrors = append(installErrors, errorMsg)
			} else { successCount++ }
		}
		fmt.Println("--- Installation Summary ---")
		fmt.Printf("Successfully installed/updated %d dependencies.\n", successCount)
		if len(installErrors) > 0 {
			fmt.Printf("%d dependencies failed to install.\n", len(installErrors))
			return modifiedProjectGom, fmt.Errorf("%d dependencies failed to install", len(installErrors))
		}
		fmt.Println("--------------------------")
	} else {
		return modifiedProjectGom, fmt.Errorf("the 'install' command takes zero or one argument ([<alias>@<version>])")
	}
	return modifiedProjectGom, nil
}

// handleGetAs installs a package via `go get` and saves an alias for it.
// Returns (modifiedGlobal bool, modifiedProject bool, error).
func handleGetAs(args []string) (bool, bool, error) {
    modifiedGlobal := false; modifiedProject := false
	getCmd := flag.NewFlagSet("get", flag.ContinueOnError)
	globalFlag := getCmd.Bool("g", false, "Save alias to global cache")
	if err := getCmd.Parse(args); err != nil {
		return false, false, fmt.Errorf("invalid flags. Usage: gom get <link> as <name>@<version> [-g]")
	}
	positionalArgs := getCmd.Args()
	if len(positionalArgs) != 3 || strings.ToLower(positionalArgs[1]) != "as" {
		return false, false, fmt.Errorf("usage: gom get <dependency_link> as <name>@<version> [-g]")
	}
	dependencyLink := positionalArgs[0]; nameAndVersion := positionalArgs[2]
	if !strings.Contains(dependencyLink, "/") { fmt.Printf("Warning: Link '%s' might not be standard.\n", dependencyLink) }
	packageName, version := parsePackageArg(nameAndVersion)
	if packageName == "" || version == "" || version == "latest" {
		return false, false, fmt.Errorf("invalid alias: '%s'. Must be <name>@<specific_version>", nameAndVersion)
	}
	fmt.Printf("-> Preparing to run: go get %s\n", dependencyLink)
	if err := runGoCommand("get", dependencyLink); err != nil {
		return false, false, fmt.Errorf("'go get %s' command failed", dependencyLink)
	}
	fmt.Printf("\n-> 'go get %s' successful.\n", dependencyLink)
	aliasVersionKey := fmt.Sprintf("%s@%s", packageName, version)
	if *globalFlag { // Save to global cache
		if err := addGlobalCachedPath(aliasVersionKey, dependencyLink); err != nil {
			return false, false, fmt.Errorf("failed to update global cache: %w", err)
		}
		modifiedGlobal = true
		fmt.Printf("-> Updated global cache: Added alias '%s' -> '%s'\n", aliasVersionKey, dependencyLink)
	} else { // Save to project gom.json
        if projectGomFile == nil || projectGomFile.filePath == "" {
             return false, false, fmt.Errorf("could not find '%s' or project root. Run 'gom init' first?", gomFileName)
        }
		if err := addProjectDependency(aliasVersionKey, dependencyLink); err != nil {
			return false, false, fmt.Errorf("failed to update project '%s': %w", gomFileName, err)
		}
		modifiedProject = true
		fmt.Printf("-> Updated project '%s': Added alias '%s' -> '%s'\n", gomFileName, aliasVersionKey, dependencyLink)
	}
	return modifiedGlobal, modifiedProject, nil
}

// handleUninstall removes an alias from the project's gom.json file.
// Returns (modifiedProjectGom bool, error).
func handleUninstall(args []string) (bool, error) {
    modifiedProjectGom := false
    if projectGomFile == nil || projectGomFile.filePath == "" {
         return modifiedProjectGom, fmt.Errorf("could not find '%s' or project root. Run 'gom init' first?", gomFileName)
    }
	if len(args) != 1 { return modifiedProjectGom, fmt.Errorf("usage: gom uninstall <alias>@<version>") }
	aliasVersion := args[0]
	packageName, version := parsePackageArg(aliasVersion)
	if packageName == "" || version == "" {
		 return modifiedProjectGom, fmt.Errorf("invalid alias format: '%s'. Must be <name>@<version>", aliasVersion)
	}
	fmt.Printf("-> Removing alias '%s' from project '%s'...\n", aliasVersion, gomFileName)
	removed, err := removeProjectDependency(aliasVersion)
	if err != nil { return modifiedProjectGom, fmt.Errorf("internal error removing dependency: %w", err) }
	if !removed { fmt.Printf("Alias '%s' not found in '%s'. Nothing to remove.\n", aliasVersion, gomFileName)
	} else {
		modifiedProjectGom = true
		fmt.Printf("Successfully removed alias '%s' from '%s'.\n", aliasVersion, gomFileName)
		fmt.Printf("Note: Run 'go mod tidy' manually if the package is no longer needed.\n")
	}
	return modifiedProjectGom, nil
}

// handleShow displays the contents of the project's gom.json file.
func handleShow(args []string) error {
    if projectGomFile == nil || projectGomFile.filePath == "" {
        return fmt.Errorf("could not find '%s' or project root. Run 'gom init' first?", gomFileName)
    }
	if len(args) > 0 { return fmt.Errorf("the 'show' command does not take any arguments") }
	// Use MarshalIndent on the whole projectGomFile struct now
	projectGomFile.mu.RLock()
	jsonData, err := json.MarshalIndent(projectGomFile, "", "  ")
	projectGomFile.mu.RUnlock()
	if err != nil { return fmt.Errorf("failed to format project data for display: %w", err) }
	fmt.Printf("--- Contents of '%s' (%s) ---\n", gomFileName, projectGomFile.filePath)
	fmt.Println(string(jsonData))
	fmt.Println("-------------------------------------------------")
	return nil
}

// --- Environment Variable Handlers ---

// handleEnvAdd adds or updates an environment variable description in gom.json.
// Usage: gom env add <VAR_NAME> --desc "Description" [--default "Default Value"]
// Returns (modifiedProjectGom bool, error)
func handleEnvAdd(args []string) (bool, error) {
	envCmd := flag.NewFlagSet("env add", flag.ContinueOnError)
	desc := envCmd.String("desc", "", "Description of the environment variable (required)")
	defVal := envCmd.String("default", "", "Optional default value")

	if err := envCmd.Parse(args); err != nil {
		return false, fmt.Errorf("invalid flags. Usage: gom env add <VAR_NAME> --desc \"...\" [--default \"...\"]")
	}
	posArgs := envCmd.Args()
	if len(posArgs) != 1 {
		return false, fmt.Errorf("usage: gom env add <VAR_NAME> --desc \"...\" [--default \"...\"]")
	}
	varName := posArgs[0]
	if *desc == "" {
		return false, fmt.Errorf("--desc flag is required")
	}
	if strings.ContainsAny(varName, "=@\"' ") { // Basic validation
		return false, fmt.Errorf("invalid environment variable name: '%s'", varName)
	}

	info := EnvironmentInfo{
		Description: *desc,
		Default:     *defVal,
	}

	if err := addProjectEnvironment(varName, info); err != nil { // From gomfile.go
		return false, err
	}
	fmt.Printf("Added/Updated environment variable '%s' in '%s'.\n", varName, gomFileName)
	return true, nil // Mark gom.json as modified
}

// handleEnvRemove removes an environment variable description from gom.json.
// Usage: gom env remove <VAR_NAME>
// Returns (modifiedProjectGom bool, error)
func handleEnvRemove(args []string) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("usage: gom env remove <VAR_NAME>")
	}
	varName := args[0]

	removed, err := removeProjectEnvironment(varName) // From gomfile.go
	if err != nil {
		return false, err
	}
	if !removed {
		fmt.Printf("Environment variable '%s' not found in '%s'.\n", varName, gomFileName)
		return false, nil // Not modified if not found
	}
	fmt.Printf("Removed environment variable '%s' from '%s'.\n", varName, gomFileName)
	return true, nil // Mark gom.json as modified
}

// handleEnvList displays documented environment variables from gom.json.
func handleEnvList(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("the 'env list' command does not take any arguments")
	}
	envs, err := getAllProjectEnvironments() // From gomfile.go
	if err != nil {
		return err
	}
	if len(envs) == 0 {
		fmt.Printf("No environment variables documented in '%s'.\n", gomFileName)
		return nil
	}

	fmt.Printf("--- Environment Variables Documented in '%s' ---\n", gomFileName)
	// Consider a more formatted output than just JSON? For now, JSON is simple.
	outputData := map[string]interface{}{"environments": envs}
	jsonData, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format environments for display: %w", err)
	}
	fmt.Println(string(jsonData))
	fmt.Println("-------------------------------------------------")
	return nil
}

// --- Cache Handlers ---

// handleCacheList displays the contents of the global alias cache file.
func handleCacheList(args []string) error {
	if len(args) > 0 { return fmt.Errorf("usage: gom cache list") }
	aliases, err := getAllGlobalCachedAliases()
	if err != nil { return fmt.Errorf("failed to read global cache: %w", err) }
	cachePath, _ := getGlobalCacheFilePath()
	if cachePath == "" { cachePath = "platform-specific cache path" }
	if len(aliases) == 0 { fmt.Printf("Global alias cache is empty (%s).\n", cachePath); return nil }

	fmt.Printf("--- Global Alias Cache Contents (%s) ---\n", cachePath)
	// Adjust structure for display if cache format changed (it didn't significantly here)
	outputData := map[string]interface{}{"aliases": aliases} // Assuming cache still stores map directly under "aliases"
	jsonData, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil { return fmt.Errorf("failed to format global cache for display: %w", err) }
	fmt.Println(string(jsonData))
	fmt.Println("-----------------------------------------")
	return nil
}

// handleCacheClear clears all entries from the global alias cache file.
func handleCacheClear(args []string) error {
	if len(args) > 0 { return fmt.Errorf("usage: gom cache clear") }
	fmt.Println("Clearing global cache...")
	if err := clearGlobalCache(); err != nil { // From cache.go
		return fmt.Errorf("failed to clear global cache: %w", err)
	}
	cachePath, _ := getGlobalCacheFilePath()
	if cachePath == "" { cachePath = "platform-specific cache path" }
	fmt.Printf("Global cache cleared (%s).\n", cachePath)
	return nil
}

// --- Doctor/Check Handlers ---

// handleCheckGo explicitly runs `go version` and updates the cache.
// Returns (modifiedGlobalCache bool, error)
func handleCheckGo(args []string) (bool, error) {
	if len(args) > 0 { return false, fmt.Errorf("usage: gom check go") }

	fmt.Println("Running `go version` to check installation and update cache...")
	versionOutput, err := runGoVersion() // From main.go (or move to utils?)
	if err != nil {
		return false, err // Error already formatted by runGoVersion
	}

	fmt.Printf("Found Go version: %s", versionOutput) // Includes newline
	err = updateGlobalCachedGoVersion(versionOutput) // Update cache
	if err != nil {
		// Log warning but don't fail the command
		fmt.Fprintf(os.Stderr, "Warning: Failed to update Go version in cache: %v\n", err)
		return false, nil // Cache wasn't successfully modified from its previous state
	}

	fmt.Println("Go version cache updated.")
	return true, nil // Cache was modified
}
