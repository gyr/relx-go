package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gyr/grxs/pkg/app"
	"github.com/gyr/grxs/pkg/config"
)

func main() {
	var configPath string
	var debugFlag bool

	flag.StringVar(&configPath, "c", "", "Path to the configuration file")
	flag.StringVar(&configPath, "config", "", "Path to the configuration file (shorthand -c)")
	flag.BoolVar(&debugFlag, "d", false, "Enable debug logging")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging (shorthand -d)")
	flag.Parse()

	// Find and load configuration
	cfgFile, err := config.FindConfigFile(configPath)
	if err != nil {
		log.Printf("Warning: %v. Proceeding without custom configuration.", err)
	}

	var cfg *config.Config
	if cfgFile != "" {
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			log.Fatalf("Error loading configuration from %s: %v", cfgFile, err)
		}
		log.Printf("Configuration loaded from %s", cfgFile)
	} else {
		// Provide a default configuration if no file is found/loaded
		cfg = &config.Config{
			CacheDir: "", // Default to empty, gitutils will use its own default
		}
	}

	// Command-line debug flag overrides config file setting
	if debugFlag {
		cfg.Debug = true
	}

	args := flag.Args() // Get non-flag arguments after flag.Parse()

	if len(args) < 1 {
		fmt.Println("Usage: gyr-grxs <command> [arguments]")
		fmt.Println("\nCommands:")
		fmt.Println("  pr <owner> <repo>       Get list of open Pull Requests (uses Gitea backend)")
		fmt.Println("  status <prj> <pkg>      Get build status for a package (uses OBS backend)")
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "pr":
		if len(commandArgs) < 2 {
			log.Fatal("Error: 'pr' requires owner and repo arguments.")
		}
		app.HandlePullRequest(cfg, commandArgs[0], commandArgs[1])

	case "status":
		if len(commandArgs) < 2 {
			log.Fatal("Error: 'status' requires project and package arguments.")
		}
		app.HandleBuildStatus(cfg, commandArgs[0], commandArgs[1])

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}