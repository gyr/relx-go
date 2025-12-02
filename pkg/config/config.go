package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"github.com/gyr/relx-go/pkg/logging"
)

// Config holds the application's configuration.
type Config struct {
	CacheDir string `yaml:"cache_dir"`
	RepoURL  string `yaml:"repo_url"`
	RepoName string
	Branch   string `yaml:"branch"`
	Logger   *logging.Logger `yaml:"-"` // Ignore logger for YAML (it's not a config value)
}

// LoadConfig loads the configuration from a YAML file.
func LoadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML config: %w", err)
	}

	// Derive RepoName from RepoURL if available
	if cfg.RepoURL != "" {
		// Remove "gitea @" prefix if present
		repoURLClean := strings.TrimPrefix(cfg.RepoURL, "gitea@")

		// Get the base name (last component) of the URL path
		base := filepath.Base(repoURLClean)
		// Remove the ".git" suffix if it exists
		cfg.RepoName = strings.TrimSuffix(base, ".git")
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

