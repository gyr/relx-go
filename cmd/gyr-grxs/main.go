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
        if cfg.Debug || debugFlag {
            log.Printf("Configuration loaded from %s", cfgFile)
        }
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

	validCommands := []string{"pr", "status", "bugowner"}

	if len(args) < 1 {
		fmt.Println("Usage: gyr-grxs <command> [arguments]")
		fmt.Println("\nCommands:")
		for _, cmd := range validCommands {
			fmt.Printf("  %s\n", cmd)
		}
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
		statusCmd := flag.NewFlagSet("status", flag.ContinueOnError)
		projectFlag := statusCmd.String("p", "", "OBS project")

		statusCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s status:\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "  -p <project>    OBS project\n")
			fmt.Fprintf(os.Stderr, "  <package>       Package\n")
		}

		err := statusCmd.Parse(commandArgs)
		if err != nil {
			if err == flag.ErrHelp {
				os.Exit(0)
			}
			// The flag package already printed an error and usage, so just exit.
			os.Exit(1)
		}

		if *projectFlag == "" || statusCmd.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "Error: 'status' requires project and package arguments.\n")
			statusCmd.Usage()
			os.Exit(1)
		}
		app.HandleBuildStatus(cfg, *projectFlag, statusCmd.Arg(0))

	case "bugowner":
		bugownerCmd := flag.NewFlagSet("bugowner", flag.ContinueOnError)
		pkgFlag := bugownerCmd.String("p", "", "Specify the package")
		maintainerFlag := bugownerCmd.String("m", "", "Specify the maintainer")

		bugownerCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s bugowner:\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "  -p <pkg>          Get bugowners for a specific package\n")
			fmt.Fprintf(os.Stderr, "  -m <maintainer>   List packages maintained by a user\n")
		}

		err = bugownerCmd.Parse(commandArgs)
		if err != nil {
			if err == flag.ErrHelp {
				os.Exit(0)
			}
			// The flag package already printed an error and usage, so just exit.
			os.Exit(1)
		}

		if (*pkgFlag != "" && *maintainerFlag != "") || (*pkgFlag == "" && *maintainerFlag == "") {
			fmt.Fprintf(os.Stderr, "Error: For 'bugowner', you must provide either -p (package) OR -m (maintainer), but not both.\n")
			bugownerCmd.Usage()
			os.Exit(1)
		}

		if *pkgFlag != "" {
			// TODO: Implement app.HandleBugownerByPackage(cfg, *pkgFlag)
			// This function should fetch bugowners for the given package.
			fmt.Printf("Getting bugowners for package: %s\n", *pkgFlag) // Placeholder
		} else { // *maintainerFlag != ""
			// TODO: Implement app.HandlePackagesByMaintainer(cfg, *maintainerFlag)
			// This function should list packages maintained by the given user.
			fmt.Printf("Listing packages maintained by: %s\n", *maintainerFlag) // Placeholder
		}

	default:
		fmt.Printf("Unknown command: %s. Possible commands are:\n", command)
		for _, cmd := range validCommands {
			fmt.Printf("  %s\n", cmd)
		}
		os.Exit(1)
	}
}
