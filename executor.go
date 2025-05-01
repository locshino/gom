// executor.go
// Contains functions for executing external commands, specifically 'go' commands.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runGoCommand executes an external 'go' command with the provided arguments.
//
// It streams the command's stdout and stderr directly to the user's console.
// It waits for the command to complete and returns any error encountered,
// typically non-nil if the command exits with a non-zero status.
//
// Example:
//   err := runGoCommand("get", "github.com/gin-gonic/gin@v1.9.1")
//   if err != nil { /* handle error */ }
func runGoCommand(goArgs ...string) error {
	fmt.Println("--- Output for 'go", strings.Join(goArgs, " "), "' ---")
	cmd := exec.Command("go", goArgs...)

	// Pipe output directly to user's terminal.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run and wait.
	err := cmd.Run()

	fmt.Println("---------------------------") // Separator.
	return err // Return command's exit status.
}
