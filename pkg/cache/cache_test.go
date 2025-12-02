package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gyr/relx-go/pkg/cache"
)

func TestNew(t *testing.T) {
	// Test with a specified baseDir
	tempDir := t.TempDir()
	c, err := cache.New(tempDir)
	if err != nil {
		t.Fatalf("New() failed with specified baseDir: %v", err)
	}
	if c == nil {
		t.Fatal("New() returned nil cache with specified baseDir")
	}
	if c.GetPath("") != tempDir {
		t.Errorf("New() baseDir mismatch: got %s, want %s", c.GetPath(""), tempDir)
	}

	// Test with empty baseDir (should return error as it's now handled by config)
	_, err = cache.New("")
	if err == nil {
		t.Fatal("New() should have failed with empty baseDir but didn't")
	}
}

func TestGetPath(t *testing.T) {
	tempDir := t.TempDir()
	c, _ := cache.New(tempDir)

	artifactName := "test_artifact"
	expectedPath := filepath.Join(tempDir, artifactName)
	gotPath := c.GetPath(artifactName)

	if gotPath != expectedPath {
		t.Errorf("GetPath() mismatch: got %s, want %s", gotPath, expectedPath)
	}
}

func TestHas(t *testing.T) {
	tempDir := t.TempDir()
	c, _ := cache.New(tempDir)

	// Test case 1: Artifact does not exist
	if c.Has("non_existent_artifact", "") {
		t.Error("Has() returned true for non-existent artifact")
	}

	// Test case 2: Artifact exists (file)
	artifactFile := filepath.Join(tempDir, "existent_file")
	err := os.WriteFile(artifactFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if !c.Has("existent_file", "") {
		t.Error("Has() returned false for existent file")
	}

	// Test case 3: Artifact exists (directory)
	artifactDir := filepath.Join(tempDir, "existent_dir")
	err = os.Mkdir(artifactDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if !c.Has("existent_dir", "") {
		t.Error("Has() returned false for existent directory")
	}

	// Test case 4: Artifact exists with inner path (e.g., .git directory check)
	gitDir := filepath.Join(artifactDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}
	if !c.Has("existent_dir", ".git") {
		t.Error("Has() returned false for existent directory with .git inner path")
	}

	// Test case 5: Artifact exists, inner path does not
	if c.Has("existent_dir", "non_existent_inner") {
		t.Error("Has() returned true for existent directory with non-existent inner path")
	}
}
