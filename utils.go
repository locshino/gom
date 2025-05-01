// utils.go
// Contains general utility functions for the gom tool.
package main

import (
	"fmt"
	"strings"
)

// parsePackageArg splits an argument string like "name@version" or just "name"
// into its name and version components. Version defaults to empty string if not present.
func parsePackageArg(arg string) (name, version string) {
	parts := strings.SplitN(arg, "@", 2)
	if len(parts) == 0 || parts[0] == "" { return "", "" } // Invalid input.
	name = parts[0]
	if len(parts) == 2 && parts[1] != "" { version = parts[1] } // Version specified.
	// Default version is ""
	return name, version
}

// printUsage displays help information about how to use the gom tool.
func printUsage() {
	// Constants are defined in config.go

	fmt.Println("Gom - Go Package Alias Manager (Local Project + Global Cache)")
	fmt.Println("\nManages Go dependencies using aliases and documents environment variables.")
	fmt.Println("\nUsage:")
	fmt.Println("  gom <command> [arguments]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  init                       Initialize an empty 'gom.json' file in the current project.")
	fmt.Println("  get <link> as <pkg>@<ver> [-g] Run 'go get <link>' and manage alias:")
	fmt.Println("                             No '-g': Add/update alias in project 'gom.json'.")
	fmt.Println("                             With '-g': Add/update alias in global cache.")
	fmt.Println("                             Version (<ver>) is required and cannot be 'latest'.")
	fmt.Println("  install [<pkg>@<ver>]     Install dependencies (alias: i)")
	fmt.Println("    gom install <pkg>@<ver>  Install specific alias (checks project then global cache).")
	fmt.Println("    gom install              Install all dependencies from project 'gom.json'.")
	fmt.Println("  uninstall <pkg>@<ver>     Remove the alias from the project's 'gom.json'.")
	fmt.Println("  show                       Display the contents of the project's 'gom.json' file.")
	fmt.Println("  env <subcommand> [args]    Manage environment variable documentation in 'gom.json'.")
	fmt.Println("    env list                 List documented environment variables.")
	fmt.Println("    env add <VAR> --desc \"..\" [--default \"..\"] Add/update an environment variable.")
	fmt.Println("    env remove <VAR>         Remove an environment variable documentation.")
	fmt.Println("  cache <subcommand>         Manage the global alias cache.")
	fmt.Println("    cache list               Display the contents of the global alias cache.")
	fmt.Println("    cache clear              Clear all entries from the global alias cache.")
	fmt.Println("  check <subcommand>         Run checks (alias: doctor).")
	fmt.Println("    check go                 Verify Go installation and update cached version.")
	fmt.Println("  version                    Display the version of the Gom tool. (Aliases: -v, --version)")


	fmt.Println("\nExamples:")
	fmt.Println("  gom init")
	fmt.Println("  gom get github.com/gin-gonic/gin@v1.9.1 as gin@v1.9.1")
	fmt.Println("  gom get gopkg.in/guregu/null.v3 as null@v3 -g")
	fmt.Println("  gom env add DB_HOST --desc \"Database host\" --default \"127.0.0.1\"")
	fmt.Println("  gom env list")
	fmt.Println("  gom install gin@v1.9.1")
	fmt.Println("  gom i null@v3")
	fmt.Println("  gom i")
	fmt.Println("  gom show")
	fmt.Println("  gom cache list")
	fmt.Println("  gom uninstall gin@v1.9.1")
	fmt.Println("  gom env remove DB_HOST")
	fmt.Println("  gom cache clear")
	fmt.Println("  gom check go")
	fmt.Println("  gom version")


	fmt.Println("\nNotes:")
	fmt.Printf("  Project-specific commands require running inside a Go module.\n")
	fmt.Printf("  Project aliases and environment info are stored in '%s' in your project root.\n", gomFileName)
	fmt.Printf("  Global alias cache is stored in the standard OS cache directory (e.g., ~/.cache/%s/%s).\n", appNameDir, globalCacheFileName)
	fmt.Printf("  This tool wraps 'go get'; actual package management relies on Go modules.\n")
}
