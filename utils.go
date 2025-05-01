// utils.go
// Contains general utility functions for the gom tool.
package main

import (
	"fmt"
	"strings"
)

// parsePackageArg splits an argument string like "name@version" or just "name"
// into its name and version components.
//
// If the input string contains '@', the part before is the name and the part after
// is the version. If '@' is not present or the part after is empty, the version
// defaults to an empty string.
// Returns empty name and version if the input is empty or invalid.
//
// Example:
//   name, version := parsePackageArg("gin@v1.9.1") // name="gin", version="v1.9.1"
//   name, version := parsePackageArg("myutil")     // name="myutil", version=""
func parsePackageArg(arg string) (name, version string) {
	parts := strings.SplitN(arg, "@", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", "" // Invalid input.
	}
	name = parts[0]
	if len(parts) == 2 && parts[1] != "" {
		version = parts[1] // Version specified.
	} else {
		version = "" // No version specified.
	}
	return name, version
}

// printUsage displays help information about how to use the gom tool.
//
// It prints a summary of the tool's purpose, usage syntax, available commands,
// examples, and notes about file locations.
func printUsage() {
	// Constants like gomFileName are defined in config.go

	fmt.Println("Gom - Go Package Alias Manager (Local Project + Global Cache)")
	fmt.Println("\nManages Go dependencies using aliases defined in 'gom.json' (project-local)")
	fmt.Println("and optionally shares aliases via a global cache.")
	fmt.Println("\nUsage:")
	fmt.Println("  gom <command> [arguments]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  init                       Initialize an empty 'gom.json' file in the current project.")
	fmt.Println("  get <link> as <pkg>@<ver> [-g] Run 'go get <link>' and manage alias:")
	fmt.Println("                             No '-g': Add/update alias '<pkg>@<ver>' -> '<link>' in project 'gom.json'.")
	fmt.Println("                             With '-g': Add/update alias '<pkg>@<ver>' -> '<link>' in global cache.")
	fmt.Println("                             Version (<ver>) is required and cannot be 'latest'.")
	fmt.Println("  install [<pkg>@<ver>]     Install dependencies (alias: i)")
	fmt.Println("    gom install <pkg>@<ver>  1. Check project 'gom.json'. 2. Check global cache.")
	fmt.Println("                             If found in global cache, add to 'gom.json'. Run 'go get <link>'.")
	fmt.Println("    gom install              Run 'go get <link>' for all dependencies in project 'gom.json'.")
	fmt.Println("  uninstall <pkg>@<ver>     Remove the alias '<pkg>@<ver>' from the project's 'gom.json'.")
	fmt.Println("                             Does not run 'go mod tidy'.")
	fmt.Println("  show                       Display the contents of the project's 'gom.json' file.")
	fmt.Println("  cache list                 Display the contents of the global alias cache.")
	fmt.Println("  cache clear                Clear all entries from the global alias cache.")
	fmt.Println("  version                    Display the version of the Gom tool. (Aliases: -v, --version)")


	fmt.Println("\nExamples:")
	fmt.Println("  gom init                      # Create ./gom.json")
	fmt.Println("  gom get github.com/gin-gonic/gin@v1.9.1 as gin@v1.9.1   # Add to ./gom.json")
	fmt.Println("  gom get gopkg.in/guregu/null.v3 as null@v3 -g           # Add to global cache")
	fmt.Println("  gom install gin@v1.9.1        # Installs gin using ./gom.json")
	fmt.Println("  gom i null@v3                 # Installs null using global cache, adds to ./gom.json")
	fmt.Println("  gom i                         # Installs gin (and null if added above) from ./gom.json")
	fmt.Println("  gom show                      # Show ./gom.json")
	fmt.Println("  gom cache list                # Show global cache")
	fmt.Println("  gom uninstall gin@v1.9.1      # Remove gin alias from ./gom.json")
	fmt.Println("  gom cache clear               # Clear global cache")
	fmt.Println("  gom version                   # Show gom tool version")


	fmt.Println("\nNotes:")
	fmt.Printf("  Project-specific commands require running inside a Go module (directory containing 'go.mod').\n")
	fmt.Printf("  Project aliases are stored in '%s' in your project root.\n", gomFileName)
	fmt.Printf("  Global cache is stored in the standard OS cache directory (e.g., ~/.cache/%s/%s).\n", appNameDir, globalCacheFileName)
	fmt.Printf("  This tool wraps 'go get'; actual package management relies on Go modules.\n")
}
