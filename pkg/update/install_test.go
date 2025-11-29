package update

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestInstaller_CreateBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.bin")
	testData := []byte("test binary")
	if err := os.WriteFile(srcFile, testData, 0755); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	installer := NewInstaller(nil)

	// Create backup
	backupFile := filepath.Join(tmpDir, "source.backup")
	if err := installer.createBackup(srcFile, backupFile); err != nil {
		t.Fatalf("createBackup() error = %v", err)
	}

	// Verify backup exists
	if _, err := os.Stat(backupFile); err != nil {
		t.Errorf("Backup file not created: %v", err)
	}

	// Verify backup content
	backupData, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupData) != string(testData) {
		t.Errorf("Backup content = %s, want %s", string(backupData), string(testData))
	}
}

func TestInstaller_CopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.bin")
	testData := []byte("test content")
	if err := os.WriteFile(srcFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	installer := NewInstaller(nil)

	// Copy file
	dstFile := filepath.Join(tmpDir, "dest.bin")
	if err := installer.copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	// Verify destination exists
	if _, err := os.Stat(dstFile); err != nil {
		t.Errorf("Destination file not created: %v", err)
	}

	// Verify content
	dstData, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstData) != string(testData) {
		t.Errorf("Destination content = %s, want %s", string(dstData), string(testData))
	}
}

func TestInstaller_VerifyBinary(t *testing.T) {
	tmpDir := t.TempDir()

	installer := NewInstaller(nil)

	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "valid binary",
			setup: func() string {
				file := filepath.Join(tmpDir, "valid.bin")
				_ = os.WriteFile(file, []byte("binary"), 0755)
				return file
			},
			wantErr: false,
		},
		{
			name: "non-executable on unix",
			setup: func() string {
				if runtime.GOOS == "windows" {
					t.Skip("Skipping non-executable test on Windows")
				}
				file := filepath.Join(tmpDir, "noexec.bin")
				_ = os.WriteFile(file, []byte("binary"), 0644)
				return file
			},
			wantErr: true,
		},
		{
			name: "directory",
			setup: func() string {
				dir := filepath.Join(tmpDir, "dir")
				_ = os.Mkdir(dir, 0755)
				return dir
			},
			wantErr: true,
		},
		{
			name: "non-existent",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			err := installer.verifyBinary(path)

			if (err != nil) != tt.wantErr {
				t.Errorf("verifyBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetExecutablePath(t *testing.T) {
	path, err := GetExecutablePath()
	if err != nil {
		t.Fatalf("GetExecutablePath() error = %v", err)
	}

	if path == "" {
		t.Error("GetExecutablePath() returned empty path")
	}

	// Verify path exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Executable path does not exist: %v", err)
	}
}

func TestCanUpdate(_ *testing.T) {
	// This test may fail in some environments due to permissions
	// We just verify that it doesn't panic
	err := CanUpdate()

	// We don't assert error here because it depends on the environment
	// Just ensure the function executes without panic
	_ = err
}

func TestNeedsSudo(t *testing.T) {
	// This test just ensures the function works
	needsSudo := NeedsSudo()

	// On Windows, should always be false
	if runtime.GOOS == "windows" && needsSudo {
		t.Error("NeedsSudo() should return false on Windows")
	}

	// We can't reliably test the Unix case without knowing the environment
	_ = needsSudo
}

func TestInstaller_Rollback(t *testing.T) {
	tmpDir := t.TempDir()

	installer := NewInstaller(nil)

	// Create current and backup files
	currentFile := filepath.Join(tmpDir, "current.bin")
	backupFile := filepath.Join(tmpDir, "current.backup")

	_ = os.WriteFile(currentFile, []byte("corrupted"), 0755)
	_ = os.WriteFile(backupFile, []byte("backup"), 0755)

	// Perform rollback
	installer.rollback(currentFile, backupFile)

	// Verify current file was restored
	data, err := os.ReadFile(currentFile)
	if err != nil {
		t.Fatalf("Failed to read current file after rollback: %v", err)
	}

	if string(data) != "backup" {
		t.Errorf("Rollback content = %s, want backup", string(data))
	}

	// Verify backup file is gone (moved to current)
	if _, err := os.Stat(backupFile); !os.IsNotExist(err) {
		t.Error("Backup file should be removed after rollback")
	}
}

func TestDefaultUpdateConfig(t *testing.T) {
	config := DefaultUpdateConfig()

	if config == nil {
		t.Fatal("DefaultUpdateConfig() returned nil")
	}

	if config.CheckInterval != 24*time.Hour {
		t.Errorf("CheckInterval = %v, want 24h", config.CheckInterval)
	}

	if config.AutoUpdate {
		t.Error("AutoUpdate should be false by default")
	}

	if !config.RequireConfirmation {
		t.Error("RequireConfirmation should be true by default")
	}

	if config.AllowPrerelease {
		t.Error("AllowPrerelease should be false by default")
	}

	if config.HTTPTimeout != 30*time.Second {
		t.Errorf("HTTPTimeout = %v, want 30s", config.HTTPTimeout)
	}
}

func TestInstaller_Install(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Install test on Windows (requires special handling)")
	}

	tmpDir := t.TempDir()

	// Create a mock current executable
	currentExe := filepath.Join(tmpDir, "current.bin")
	if err := os.WriteFile(currentExe, []byte("old version"), 0755); err != nil {
		t.Fatalf("Failed to create current executable: %v", err)
	}

	// Create a new version binary
	newExe := filepath.Join(tmpDir, "new.bin")
	if err := os.WriteFile(newExe, []byte("new version"), 0755); err != nil {
		t.Fatalf("Failed to create new executable: %v", err)
	}

	installer := NewInstaller(nil)

	// Note: This test is limited as we can't actually replace the running executable
	// But we can test the backup and copy logic
	err := installer.createBackup(currentExe, currentExe+".backup")
	if err != nil {
		t.Errorf("createBackup() error = %v", err)
	}

	// Verify backup was created
	if _, err := os.Stat(currentExe + ".backup"); err != nil {
		t.Error("Backup file should exist")
	}
}

func TestInstaller_ReplaceFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping ReplaceFile test on Windows")
	}

	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.bin")
	if err := os.WriteFile(srcFile, []byte("new content"), 0755); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create destination file
	dstFile := filepath.Join(tmpDir, "dest.bin")
	if err := os.WriteFile(dstFile, []byte("old content"), 0755); err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}

	installer := NewInstaller(nil)

	// Replace file
	err := installer.replaceFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("replaceFile() error = %v", err)
	}

	// Verify destination was replaced
	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}

	if string(data) != "new content" {
		t.Errorf("Destination content = %s, want 'new content'", string(data))
	}
}

func TestInstaller_ReplaceFile_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.bin")
	if err := os.WriteFile(srcFile, []byte("new content"), 0755); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create destination file
	dstFile := filepath.Join(tmpDir, "dest.bin")
	if err := os.WriteFile(dstFile, []byte("old content"), 0755); err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}

	installer := NewInstaller(nil)

	// Replace file using Windows method
	err := installer.replaceFileWindows(srcFile, dstFile)
	if err != nil {
		t.Fatalf("replaceFileWindows() error = %v", err)
	}

	// Verify destination was replaced
	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}

	if string(data) != "new content" {
		t.Errorf("Destination content = %s, want 'new content'", string(data))
	}
}

func TestInstaller_CreateBackup_Error(t *testing.T) {
	tmpDir := t.TempDir()

	installer := NewInstaller(nil)

	// Try to backup non-existent file
	err := installer.createBackup("/nonexistent/file", tmpDir+"/backup")
	if err == nil {
		t.Error("createBackup() should return error for non-existent file")
	}
}

func TestInstaller_CopyFile_Error(t *testing.T) {
	tmpDir := t.TempDir()

	installer := NewInstaller(nil)

	// Try to copy non-existent file
	err := installer.copyFile("/nonexistent/file", tmpDir+"/dest")
	if err == nil {
		t.Error("copyFile() should return error for non-existent file")
	}
}
