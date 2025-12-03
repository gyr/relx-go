// Package command provides a generic interface for running external commands.
// This abstraction allows for easier testing by enabling the mocking of command execution.
package command

import (
	"context"
	"os/exec"
)

// Runner defines a standard interface for executing external commands.
// By depending on this interface rather than directly on the 'os/exec' package,
// application logic can be unit-tested with mock implementations.
type Runner interface {
	// Run executes the specified command with the given arguments.
	// It uses a context for timeout and cancellation control.
	// It returns the combined stdout and stderr output, or an error if the command fails.
	Run(ctx context.Context, workDir, name string, args ...string) ([]byte, error)
}

// DefaultRunner is the default implementation of the Runner interface.
// It uses the 'os/exec' package to run commands on the host system.
type DefaultRunner struct{}

// Run executes a command using exec.CommandContext, which respects the context's deadline.
func (r *DefaultRunner) Run(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	// cmd.CombinedOutput() runs the command and returns the combined standard output and standard error.
	return cmd.CombinedOutput()
}
