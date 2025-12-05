// Package command provides a generic interface for running external commands.
// This abstraction allows for easier testing by enabling the mocking of command execution.
package command

import (
	"context"
	"os"
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
	// RunInteractive executes the specified command in interactive mode.
	RunInteractive(ctx context.Context, workDir, name string, args ...string) error
	// RunPipeline executes a pipeline of two commands.
	RunPipeline(ctx context.Context, workDir string, cmd1, cmd2 []string) error
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

// RunInteractive executes a command in interactive mode.
func (r *DefaultRunner) RunInteractive(ctx context.Context, workDir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunPipeline executes a pipeline of two commands.
func (r *DefaultRunner) RunPipeline(ctx context.Context, workDir string, cmd1Args, cmd2Args []string) error {
	cmd1 := exec.CommandContext(ctx, cmd1Args[0], cmd1Args[1:]...)
	cmd2 := exec.CommandContext(ctx, cmd2Args[0], cmd2Args[1:]...)

	if workDir != "" {
		cmd1.Dir = workDir
		cmd2.Dir = workDir
	}

	pipe, err := cmd1.StdoutPipe()
	if err != nil {
		return err
	}
	cmd2.Stdin = pipe
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr

	if err := cmd1.Start(); err != nil {
		return err
	}
	if err := cmd2.Start(); err != nil {
		return err
	}

	if err := cmd1.Wait(); err != nil {
		return err
	}
	if err := cmd2.Wait(); err != nil {
		return err
	}

	return nil
}
