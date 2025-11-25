package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Installer handles installing updates.
type Installer struct {
	config *UpdateConfig
}

// NewInstaller creates a new installer.
func NewInstaller(config *UpdateConfig) *Installer {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	return &Installer{
		config: config,
	}
}

// Install installs an update by replacing the current binary.
// This operation is atomic - either it succeeds completely or rolls back.
func (i *Installer) Install(downloadedPath string) error {
	// Get current executable path
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Resolve symlinks
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Get file info to preserve permissions
	fileInfo, err := os.Stat(currentPath)
	if err != nil {
		return fmt.Errorf("failed to stat current executable: %w", err)
	}

	// Create backup
	backupPath := currentPath + ".backup"
	if err := i.createBackup(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Make downloaded file executable
	if err := os.Chmod(downloadedPath, fileInfo.Mode()); err != nil {
		i.rollback(currentPath, backupPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Replace the binary atomically
	if err := i.replaceFile(downloadedPath, currentPath); err != nil {
		i.rollback(currentPath, backupPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Verify the new binary
	if err := i.verifyBinary(currentPath); err != nil {
		i.rollback(currentPath, backupPath)
		return fmt.Errorf("verification failed: %w", err)
	}

	// Remove backup on success
	os.Remove(backupPath)

	return nil
}

// createBackup creates a backup of the current binary.
func (i *Installer) createBackup(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}

	// Write backup
	if err := os.WriteFile(dst, data, 0755); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// replaceFile replaces the target file with the source file atomically.
func (i *Installer) replaceFile(src, dst string) error {
	// On Unix-like systems, we can use rename for atomic replacement
	// On Windows, we need to remove the target first
	if runtime.GOOS == "windows" {
		// On Windows, we can't replace a running executable
		// We need to use a different strategy
		return i.replaceFileWindows(src, dst)
	}

	// On Unix, rename is atomic if on the same filesystem
	// First, move to same directory to ensure atomic operation
	tmpPath := dst + ".new"

	// Copy to temp location in same directory
	if err := i.copyFile(src, tmpPath); err != nil {
		return fmt.Errorf("failed to copy to temp location: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, dst); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}

// replaceFileWindows handles file replacement on Windows.
func (i *Installer) replaceFileWindows(src, dst string) error {
	// On Windows, we can't replace a running executable
	// We rename the current executable and copy the new one
	// The old file will be deleted on next boot or next update

	// Rename current executable
	oldPath := dst + ".old"
	if err := os.Rename(dst, oldPath); err != nil {
		return fmt.Errorf("failed to rename current executable: %w", err)
	}

	// Copy new executable
	if err := i.copyFile(src, dst); err != nil {
		// Try to restore
		os.Rename(oldPath, dst)
		return fmt.Errorf("failed to copy new executable: %w", err)
	}

	// Mark old file for deletion
	os.Remove(oldPath)

	return nil
}

// copyFile copies a file from src to dst.
func (i *Installer) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}

	if err := os.WriteFile(dst, data, 0755); err != nil {
		return fmt.Errorf("failed to write destination: %w", err)
	}

	return nil
}

// verifyBinary verifies that the new binary is valid and executable.
func (i *Installer) verifyBinary(path string) error {
	// Check if file exists and is executable
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	// On Unix, check execute permission
	if runtime.GOOS != "windows" {
		mode := info.Mode()
		if mode&0111 == 0 {
			return fmt.Errorf("binary is not executable")
		}
	}

	// Try to execute the binary with --version flag
	// This is a basic sanity check
	cmd := exec.Command(path, "--version")
	if err := cmd.Run(); err != nil {
		// Don't fail if --version doesn't work, as it might not be implemented
		// Just warn
		fmt.Fprintf(os.Stderr, "Warning: binary verification returned error: %v\n", err)
	}

	return nil
}

// rollback rolls back to the backup.
func (i *Installer) rollback(currentPath, backupPath string) {
	if _, err := os.Stat(backupPath); err != nil {
		return // No backup to restore
	}

	// Remove current (possibly corrupted) file
	os.Remove(currentPath)

	// Restore backup
	if err := os.Rename(backupPath, currentPath); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to rollback: %v\n", err)
		fmt.Fprintf(os.Stderr, "Backup is at: %s\n", backupPath)
	}
}

// InstallAndRestart installs the update and restarts the application.
func (i *Installer) InstallAndRestart(downloadedPath string, args []string) error {
	// Get current executable path
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Install the update
	if err := i.Install(downloadedPath); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Restart the application
	cmd := exec.Command(currentPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to restart: %w", err)
	}

	// Exit current process
	os.Exit(0)

	return nil
}

// GetExecutablePath returns the path of the current executable.
func GetExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	return path, nil
}

// CanUpdate checks if the current process has permission to update the binary.
func CanUpdate() error {
	path, err := GetExecutablePath()
	if err != nil {
		return err
	}

	// Check if we can write to the executable
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("cannot write to executable (try running with sudo): %w", err)
	}
	file.Close()

	return nil
}

// NeedsSudo returns true if sudo is required to update the binary.
func NeedsSudo() bool {
	err := CanUpdate()
	return err != nil && runtime.GOOS != "windows"
}
