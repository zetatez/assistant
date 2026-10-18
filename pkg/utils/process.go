package utils

import (
	"fmt"
	"os/exec"
)

// ProcessRunning reports whether a process with the exact name is running.
// pgrep is available on both Arch Linux and macOS.
func ProcessRunning(name string) bool {
	if name == "" {
		return false
	}
	return exec.Command("pgrep", "-x", name).Run() == nil
}

// KillProcess terminates every process with the exact name.
func KillProcess(name string) error {
	if name == "" {
		return fmt.Errorf("proc cannot be empty")
	}
	return exec.Command("pkill", name).Run()
}
