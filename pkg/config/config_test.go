package config_test

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

func TestLoadConfigTildeExpansion(t *testing.T) {
	var err error
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Get current user for home dir path
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	// Write config content with a tilde path
	configContent := []byte("cache_dir: \"~/.test-cache\"")
	err = os.WriteFile(configFile, configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Load the config
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	cfg.Logger = logging.NewLogger(logging.LevelDebug)

	// Check if the tilde was expanded
	expectedPath := filepath.Join(currentUser.HomeDir, ".test-cache")
	if cfg.CacheDir != expectedPath {
		t.Errorf("Expected CacheDir to be %q, but got %q", expectedPath, cfg.CacheDir)
	}
}

func TestLoadConfigDefaultCacheDir(t *testing.T) {
	var err error
	// Create an empty temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := []byte("debug: true") // A file with no cache_dir
	err = os.WriteFile(configFile, configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Get current user for home dir path
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	// Load the config
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	cfg.Logger = logging.NewLogger(logging.LevelDebug)

	// Check if the default cache dir is set correctly
	expectedPath := filepath.Join(currentUser.HomeDir, ".cache", "relx-go")
	if cfg.CacheDir != expectedPath {
		t.Errorf("Expected default CacheDir to be %q, but got %q", expectedPath, cfg.CacheDir)
	}
}
