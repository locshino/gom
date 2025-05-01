// gomfile.go
// Manages the project-local gom.json file for dependency aliases.
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

// GomFile represents the structure of the gom.json file.
type GomFile struct {
	Dependencies map[string]string `json:"dependencies"` // Map: alias@version -> import_path
	filePath     string            // Absolute path to the gom.json file.
	mu           sync.RWMutex      // Protects concurrent access to Dependencies map.
}

// projectGomFile holds the loaded gom.json data for the current project.
// It's initialized by initProjectGomFile.
var projectGomFile *GomFile

// findProjectRoot searches upwards from the current directory for a go.mod file.
//
// It iterates up the directory tree until go.mod is found or the root is reached.
// Returns the absolute path to the directory containing go.mod.
// Returns an error if go.mod is not found or if there's a filesystem error.
func findProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	}

	dir := currentDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil // Found.
		} else if !errors.Is(err, os.ErrNotExist) {
			// Filesystem error other than "not found".
			return "", fmt.Errorf("error checking for go.mod in '%s': %w", dir, err)
		}

		// Move up.
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached root.
			return "", fmt.Errorf("could not find go.mod in the current directory or any parent directory. Please run 'gom init' or 'gom' commands inside a Go module")
		}
		dir = parentDir
	}
}

// initProjectGomFile loads the gom.json file from the specified project root.
//
// It reads the file if it exists and unmarshals the JSON data into the global
// projectGomFile variable. If the file doesn't exist, it initializes an empty
// projectGomFile. It handles potential JSON parsing errors.
// Assumes projectRoot is a valid directory path.
// Returns an error only if the file exists but is unreadable or contains invalid JSON.
func initProjectGomFile(projectRoot string) error {
	gomFilePath := filepath.Join(projectRoot, gomFileName)
	projectGomFile = &GomFile{
		Dependencies: make(map[string]string),
		filePath:     gomFilePath,
	}

	// Read file content.
	data, err := os.ReadFile(gomFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // File not existing is okay.
		}
		return fmt.Errorf("could not read existing '%s' file at '%s': %w", gomFileName, gomFilePath, err)
	}

	// Parse JSON data.
	projectGomFile.mu.Lock() // Lock for writing to the struct.
	defer projectGomFile.mu.Unlock()

	if err := json.Unmarshal(data, &projectGomFile); err != nil {
		fmt.Printf("Warning: Could not parse existing '%s' file at '%s'. Error: %v\n", gomFileName, gomFilePath, err)
		projectGomFile.Dependencies = make(map[string]string) // Reset on error.
		return fmt.Errorf("invalid JSON format in '%s'", gomFilePath) // Indicate bad format.
	}

	// Ensure map is initialized.
	if projectGomFile.Dependencies == nil {
		projectGomFile.Dependencies = make(map[string]string)
	}
	return nil
}

// saveProjectGomFile writes the current in-memory dependency data to the project's gom.json file.
//
// It marshals the projectGomFile.Dependencies map into JSON format and writes it
// to the filePath stored in projectGomFile. It handles file writing errors.
// Does nothing if projectGomFile or its filePath is not set (e.g., outside a project).
func saveProjectGomFile() error {
	if projectGomFile == nil || projectGomFile.filePath == "" {
		return nil // Nothing to save.
	}

	projectGomFile.mu.RLock() // Read lock for accessing Dependencies.
	saveData := map[string]interface{}{"dependencies": projectGomFile.Dependencies}
	data, err := json.MarshalIndent(saveData, "", "  ")
	projectGomFile.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("could not marshal project dependencies to JSON: %w", err)
	}

	projectGomFile.mu.Lock() // Lock for writing file.
	defer projectGomFile.mu.Unlock()

	// Write file (permissions 0644: rw-r--r--).
	if err := os.WriteFile(projectGomFile.filePath, data, 0644); err != nil {
		return fmt.Errorf("could not write to project '%s' file at '%s': %w", gomFileName, projectGomFile.filePath, err)
	}
	return nil
}

// addProjectDependency adds or updates a dependency alias in the in-memory projectGomFile.
//
// It takes the alias key (e.g., "name@version") and the import path.
// It ensures the internal map exists and updates the entry.
// Returns an error if projectGomFile is not initialized.
// Note: This function does not save the changes to disk.
func addProjectDependency(aliasVersion, importPath string) error {
    if projectGomFile == nil {
        return fmt.Errorf("internal error: projectGomFile not initialized before adding dependency")
    }
	projectGomFile.mu.Lock() // Lock for writing to map.
	defer projectGomFile.mu.Unlock()

	if projectGomFile.Dependencies == nil { // Should be initialized, but double-check.
		projectGomFile.Dependencies = make(map[string]string)
	}
	projectGomFile.Dependencies[aliasVersion] = importPath
	return nil
}

// removeProjectDependency removes a dependency alias from the in-memory projectGomFile.
//
// It takes the alias key (e.g., "name@version").
// Returns true if the key existed and was removed, false otherwise.
// Returns an error if projectGomFile is not initialized.
// Note: This function does not save the changes to disk.
func removeProjectDependency(aliasVersion string) (bool, error) {
    if projectGomFile == nil {
        return false, fmt.Errorf("internal error: projectGomFile not initialized before removing dependency")
    }
    projectGomFile.mu.Lock() // Lock for writing to map.
    defer projectGomFile.mu.Unlock()

    if projectGomFile.Dependencies == nil {
        return false, nil // Key doesn't exist.
    }

    _, exists := projectGomFile.Dependencies[aliasVersion]
    if exists {
        delete(projectGomFile.Dependencies, aliasVersion)
        return true, nil // Removed successfully.
    }
    return false, nil // Key didn't exist.
}


// getProjectDependency retrieves the import path for a specific alias@version from the project data.
//
// Returns the path and true if found, otherwise empty string and false.
// Returns false if projectGomFile is not initialized or the alias is not found.
func getProjectDependency(aliasVersion string) (string, bool) {
    if projectGomFile == nil {
        return "", false
    }
	projectGomFile.mu.RLock() // Read lock.
	defer projectGomFile.mu.RUnlock()

    if projectGomFile.Dependencies == nil {
        return "", false
    }
	path, found := projectGomFile.Dependencies[aliasVersion]
	return path, found && path != "" // Consider empty path as not found.
}

// getAllProjectDependencies returns a copy of the current project dependencies map.
//
// Returns a copy to prevent external modification of the internal map.
// Returns an error if projectGomFile is not initialized or if not inside a Go project.
func getAllProjectDependencies() (map[string]string, error) {
    if projectGomFile == nil {
         return nil, fmt.Errorf("internal error: projectGomFile not initialized before getting all dependencies")
    }
     if projectGomFile.filePath == "" {
        // Not inside a Go project context.
        return nil, fmt.Errorf("cannot list dependencies: not inside a Go project (go.mod not found)")
    }

    projectGomFile.mu.RLock() // Read lock.
    defer projectGomFile.mu.RUnlock()

    if projectGomFile.Dependencies == nil {
         return make(map[string]string), nil // Return empty map.
    }
    // Create and return a copy.
    depsCopy := make(map[string]string, len(projectGomFile.Dependencies))
    for k, v := range projectGomFile.Dependencies {
        depsCopy[k] = v
    }
    return depsCopy, nil
}
