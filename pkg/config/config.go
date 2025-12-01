package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"

	"github.com/gyr/relx-go/pkg/logging"
)

// Config holds the application's configuration.
type Config struct {
	CacheDir string
	RepoURL  string
	RepoName string
	Branch   string
	Logger   *logging.Logger
}

// LoadConfig loads the configuration from a Lua file.
func LoadConfig(configPath string) (*Config, error) {
	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile(configPath); err != nil {
		return nil, fmt.Errorf("failed to execute Lua config file: %w", err)
	}

	// The Lua config file is expected to return a table named 'config'
	if tbl := L.Get(-1); tbl.Type() == lua.LTTable {
		luaConfig := tbl.(*lua.LTable)

		cfg := &Config{}

		// Read cache_dir
		if cacheDir := luaConfig.RawGetString("cache_dir"); cacheDir.Type() == lua.LTString {
			cfg.CacheDir = cacheDir.String()
		}

		// Read repo_url
		if repoURL := luaConfig.RawGetString("repo_url"); repoURL.Type() == lua.LTString {
			cfg.RepoURL = repoURL.String()

			// Derive RepoName from RepoURL
			if cfg.RepoURL != "" {
				// Remove "gitea @" prefix if present
				repoURLClean := strings.TrimPrefix(cfg.RepoURL, "gitea@")

				// Get the base name (last component) of the URL path
				base := filepath.Base(repoURLClean)
				// Remove the ".git" suffix if it exists
				cfg.RepoName = strings.TrimSuffix(base, ".git")
			}
		}

		// Read branch
		if branch := luaConfig.RawGetString("branch"); branch.Type() == lua.LTString {
			cfg.Branch = branch.String()
		}

		return cfg, nil
	}

	return nil, fmt.Errorf("lua config file did not return a table")
}

// FindConfigFile searches for the configuration file in a predefined order.
func FindConfigFile(cliConfigPath string) (string, error) {
	// 1. Check command-line flag path
	if cliConfigPath != "" {
		if _, err := os.Stat(cliConfigPath); err == nil {
			return cliConfigPath, nil
		}
	}

	// 2. Check environment variable REXS_GO_CONFIG_FILE
	if envPath := os.Getenv("REXS_GO_CONFIG_FILE"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
	}

	// 3. Check ~/.config/rexs-go/config.lua
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("could not get current user: %w", err)
	}
	homeConfigPath := filepath.Join(currentUser.HomeDir, ".config", "rexs-go", "config.lua")
	if _, err := os.Stat(homeConfigPath); err == nil {
		return homeConfigPath, nil
	}

	// 4. Check /etc/rexs-go/config.lua
	etcConfigPath := filepath.Join("/etc", "rexs-go", "config.lua")
	if _, err := os.Stat(etcConfigPath); err == nil {
		return etcConfigPath, nil
	}

	return "", fmt.Errorf("no configuration file found")
}
