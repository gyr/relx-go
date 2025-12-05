package config

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"github.com/gyr/relx-go/pkg/logging"
)

// PackageFilter defines the structure for a package filter, associating
// a pattern with a specific repository.
type PackageFilter struct {
	Pattern    string `yaml:"pattern"`
	Repository string `yaml:"repository"`
}

// Config holds the application's configuration.
type Config struct {
	CacheDir                string          `yaml:"cache_dir"`
	RepoURL                 string          `yaml:"repo_url"`
	RepoBranch              string          `yaml:"repo_branch"`
	OBSAPIURL               string          `yaml:"obs_api_url"`
	PRReviewer              string          `yaml:"pr_reviewer"`
	PackageFilterPatterns   []PackageFilter `yaml:"package_filter_patterns"`
	BinaryFilterPatterns    []string        `yaml:"binary_filter_patterns"`
	OperationTimeoutSeconds int             `yaml:"operation_timeout_seconds"` // Timeout for various operations in seconds
	Logger                  *logging.Logger `yaml:"-"`                         // Ignore logger for YAML (it's not a config value)
	OutputWriter            io.Writer       `yaml:"-"`                         // Ignore output writer for YAML (it's not a config value)
}

// LoadConfig loads the configuration from a YAML file.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML config: %w", err)
	}

	// Set default OperationTimeoutSeconds if not provided
	if cfg.OperationTimeoutSeconds == 0 {
		cfg.OperationTimeoutSeconds = 300 // Default to 5 minutes
	}

	// Set default CacheDir if not provided in config file
	if cfg.CacheDir == "" {
		currentUser, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("config: could not get current user: %w", err)
		}
		cfg.CacheDir = filepath.Join(currentUser.HomeDir, ".cache", "relx-go")
	}

	// Expand tilde in CacheDir path
	if strings.HasPrefix(cfg.CacheDir, "~") {
		currentUser, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("config: could not get current user to expand tilde in cache_dir: %w", err)
		}
		cfg.CacheDir = filepath.Join(currentUser.HomeDir, cfg.CacheDir[1:])
	}

	return cfg, nil
}

// FindConfigFile searches for the configuration file in a predefined order.
func FindConfigFile(cliConfigPath string) (string, error) {
	// 1. Check command-line flag path
	if cliConfigPath != "" {
		if _, err := os.Stat(cliConfigPath); err == nil {
			return cliConfigPath, nil
		}
	}

	// 2. Check environment variable RELX_GO_CONFIG_FILE
	if envPath := os.Getenv("RELX_GO_CONFIG_FILE"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
	}

	// 3. Check ~/.config/relx-go/config.yaml
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("could not get current user: %w", err)
	}
	homeConfigPath := filepath.Join(currentUser.HomeDir, ".config", "relx-go", "config.yaml")
	if _, err := os.Stat(homeConfigPath); err == nil {
		return homeConfigPath, nil
	}

	// 4. Check /etc/relx-go/config.yaml
	etcConfigPath := filepath.Join("/etc", "relx-go", "config.yaml")
	if _, err := os.Stat(etcConfigPath); err == nil {
		return etcConfigPath, nil
	}

	return "", fmt.Errorf("no configuration file found")
}
