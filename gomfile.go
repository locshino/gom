// gomfile.go
// Manages the project-local gom.json file (aliases and environment info).
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// gomFileName is defined in config.go

// EnvironmentInfo holds details about a configured environment variable.
type EnvironmentInfo struct {
	Default     string `json:"default,omitempty"` // Optional default value.
	Description string `json:"description"`       // Required description.
}

// GomFile represents the structure of the gom.json file.
type GomFile struct {
	// Dependency aliases: alias@version -> import_path
	Dependencies map[string]string `json:"dependencies,omitempty"`
	// Environment variable documentation: VAR_NAME -> EnvironmentInfo
	Environments map[string]EnvironmentInfo `json:"environments,omitempty"`
	filePath     string                     // Absolute path to the gom.json file.
	mu           sync.RWMutex               // Protects concurrent access.
}

// projectGomFile holds the loaded gom.json data for the current project.
var projectGomFile *GomFile

// findProjectRoot searches upwards for go.mod.
func findProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil { return "", fmt.Errorf("could not get current working directory: %w", err) }
	dir := currentDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil { return dir, nil } // Found.
		if !errors.Is(err, os.ErrNotExist) { return "", fmt.Errorf("error checking for go.mod in '%s': %w", dir, err) }
		parentDir := filepath.Dir(dir)
		if parentDir == dir { return "", fmt.Errorf("could not find go.mod. Please run 'gom' commands inside a Go module") }
		dir = parentDir
	}
}

// initProjectGomFile loads gom.json from the project root.
// Initializes an empty struct if the file doesn't exist.
func initProjectGomFile(projectRoot string) error {
	gomFilePath := filepath.Join(projectRoot, gomFileName)
	projectGomFile = &GomFile{ // Initialize maps immediately
		Dependencies: make(map[string]string),
		Environments: make(map[string]EnvironmentInfo),
		filePath:     gomFilePath,
	}

	data, err := os.ReadFile(gomFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { return nil } // File not existing is okay.
		return fmt.Errorf("could not read existing '%s' file at '%s': %w", gomFileName, gomFilePath, err)
	}

	projectGomFile.mu.Lock()
	defer projectGomFile.mu.Unlock()
	if err := json.Unmarshal(data, &projectGomFile); err != nil {
		fmt.Printf("Warning: Could not parse existing '%s' file at '%s'. Error: %v\n", gomFileName, gomFilePath, err)
		// Reset maps on parse error.
		projectGomFile.Dependencies = make(map[string]string)
		projectGomFile.Environments = make(map[string]EnvironmentInfo)
		return fmt.Errorf("invalid JSON format in '%s'", gomFilePath)
	}

	// Ensure maps are initialized after unmarshaling (if JSON had null).
	if projectGomFile.Dependencies == nil { projectGomFile.Dependencies = make(map[string]string) }
	if projectGomFile.Environments == nil { projectGomFile.Environments = make(map[string]EnvironmentInfo) }
	return nil
}

// saveProjectGomFile writes the current data to the project's gom.json file.
func saveProjectGomFile() error {
	if projectGomFile == nil || projectGomFile.filePath == "" { return nil } // Nothing to save.

	projectGomFile.mu.RLock()
	// Marshal the entire GomFile struct.
	data, err := json.MarshalIndent(projectGomFile, "", "  ")
	projectGomFile.mu.RUnlock()
	if err != nil { return fmt.Errorf("could not marshal project data to JSON: %w", err) }

	projectGomFile.mu.Lock()
	defer projectGomFile.mu.Unlock()
	if err := os.WriteFile(projectGomFile.filePath, data, 0644); err != nil { // Perms 0644: rw-r--r--.
		return fmt.Errorf("could not write to project '%s' file at '%s': %w", gomFileName, projectGomFile.filePath, err)
	}
	return nil
}

// --- Dependency Alias Management ---

func addProjectDependency(aliasVersion, importPath string) error {
    if projectGomFile == nil { return fmt.Errorf("internal error: projectGomFile not initialized") }
	projectGomFile.mu.Lock()
	defer projectGomFile.mu.Unlock()
	if projectGomFile.Dependencies == nil { projectGomFile.Dependencies = make(map[string]string) }
	projectGomFile.Dependencies[aliasVersion] = importPath
	return nil
}

func removeProjectDependency(aliasVersion string) (bool, error) {
    if projectGomFile == nil { return false, fmt.Errorf("internal error: projectGomFile not initialized") }
    projectGomFile.mu.Lock()
    defer projectGomFile.mu.Unlock()
    if projectGomFile.Dependencies == nil { return false, nil } // Not found if map nil.
    _, exists := projectGomFile.Dependencies[aliasVersion]
    if exists { delete(projectGomFile.Dependencies, aliasVersion); return true, nil }
    return false, nil // Not found.
}

func getProjectDependency(aliasVersion string) (string, bool) {
    if projectGomFile == nil { return "", false }
	projectGomFile.mu.RLock()
	defer projectGomFile.mu.RUnlock()
    if projectGomFile.Dependencies == nil { return "", false }
	path, found := projectGomFile.Dependencies[aliasVersion]
	return path, found && path != ""
}

func getAllProjectDependencies() (map[string]string, error) {
    if projectGomFile == nil { return nil, fmt.Errorf("internal error: projectGomFile not initialized") }
    if projectGomFile.filePath == "" { return nil, fmt.Errorf("cannot list dependencies: not inside a Go project") }
    projectGomFile.mu.RLock()
    defer projectGomFile.mu.RUnlock()
    if projectGomFile.Dependencies == nil { return make(map[string]string), nil }
    depsCopy := make(map[string]string, len(projectGomFile.Dependencies))
    for k, v := range projectGomFile.Dependencies { depsCopy[k] = v }
    return depsCopy, nil
}

// --- Environment Variable Management ---

// addProjectEnvironment adds or updates an environment variable's info in gom.json.
func addProjectEnvironment(name string, info EnvironmentInfo) error {
	if projectGomFile == nil { return fmt.Errorf("internal error: projectGomFile not initialized") }
	projectGomFile.mu.Lock()
	defer projectGomFile.mu.Unlock()
	if projectGomFile.Environments == nil { projectGomFile.Environments = make(map[string]EnvironmentInfo) }
	projectGomFile.Environments[name] = info
	return nil
}

// removeProjectEnvironment removes an environment variable's info from gom.json.
// Returns true if removed, false if not found.
func removeProjectEnvironment(name string) (bool, error) {
	if projectGomFile == nil { return false, fmt.Errorf("internal error: projectGomFile not initialized") }
	projectGomFile.mu.Lock()
	defer projectGomFile.mu.Unlock()
	if projectGomFile.Environments == nil { return false, nil } // Not found.
	_, exists := projectGomFile.Environments[name]
	if exists { delete(projectGomFile.Environments, name); return true, nil }
	return false, nil // Not found.
}

// getAllProjectEnvironments returns a copy of the environments map.
func getAllProjectEnvironments() (map[string]EnvironmentInfo, error) {
	if projectGomFile == nil { return nil, fmt.Errorf("internal error: projectGomFile not initialized") }
	if projectGomFile.filePath == "" { return nil, fmt.Errorf("cannot list environments: not inside a Go project") }
	projectGomFile.mu.RLock()
	defer projectGomFile.mu.RUnlock()
	if projectGomFile.Environments == nil { return make(map[string]EnvironmentInfo), nil }
	envsCopy := make(map[string]EnvironmentInfo, len(projectGomFile.Environments))
	for k, v := range projectGomFile.Environments { envsCopy[k] = v }
	return envsCopy, nil
}
