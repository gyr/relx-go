package main

import (
	"context" // Import context for cancellation and timeouts
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gyr/relx-go/pkg/app"
	"github.com/gyr/relx-go/pkg/command" // Import the new command runner
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

func main() {
	var verbose, debug bool
	var configPath string

	flag.BoolVar(&verbose, "v", false, "Enable verbose output (INFO level).")
	flag.BoolVar(&debug, "d", false, "Enable debug output (DEBUG level).")
	flag.StringVar(&configPath, "c", "", "Path to the configuration file.")
	flag.Parse()

	var logLevel logging.LogLevel
	if debug {
		logLevel = logging.LevelDebug
	} else if verbose {
		logLevel = logging.LevelInfo
	} else {
		logLevel = logging.LevelError
	}
	logger := logging.NewLogger(logLevel)

	// Find and load configuration
	cfgFile, err := config.FindConfigFile(configPath)
	if err != nil {
		logger.Infof("Warning: %v. Proceeding without custom configuration.", err)
	}

	var cfg *config.Config
	if cfgFile != "" {
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			logger.Fatalf("Error loading configuration from %s: %v", cfgFile, err)
		}
		cfg.Logger = logger // Assign the logger to the config
		cfg.OutputWriter = os.Stdout
		logger.Debug("Configuration loaded from: ", cfgFile)
	} else {
		// Provide a default configuration if no file is found/loaded
		cfg = &config.Config{
			CacheDir:     "",        // Default to empty, gitutils will use its own default
			Logger:       logger,    // Assign the logger to the default config
			OutputWriter: os.Stdout, // Default to os.Stdout for application output
			// OperationTimeoutSeconds will be set to default 300 in config.LoadConfig if not specified
		}
	}

	// Initialize the default command runner and the root context for the application.
	// The runner is passed down to functions that need to execute external commands,
	// enabling dependency injection for easier testing.
	defaultRunner := &command.DefaultRunner{}
	ctx := context.Background() // Use context.Background() as the root context

	args := flag.Args() // Get non-flag arguments after flag.Parse()

	validCommands := []string{"review", "bugowner", "artifact"}

	if len(args) < 1 {
		fmt.Println("Usage: relx-go <command> [arguments]")
		fmt.Println("\nCommands:")
		for _, cmd := range validCommands {
			fmt.Printf("  %s\n", cmd)
		}
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "review":
		reviewCmd := flag.NewFlagSet("review", flag.ContinueOnError)
		branchFlag := reviewCmd.String("b", "", "Specify the branch")
		prIDFlag := reviewCmd.String("p", "", "Specify one or more comma-separated PR IDs")
		repoFlag := reviewCmd.String("r", "", "Specify the repository")
		userFlag := reviewCmd.String("u", "", "Specify the PR reviewer")

		reviewCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s review:\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "  -b, --branch <branch>       Get pull requests for a specific branch (mandatory)\n")
			fmt.Fprintf(os.Stderr, "  -p, --pr-id <id1,id2,...> Filter pull requests by specific PR IDs (optional)\n")
			fmt.Fprintf(os.Stderr, "  -r, --repository <repository>   Get pull requests for a specific repository\n")
			fmt.Fprintf(os.Stderr, "  -u, --user <user>             Specify the PR reviewer\n")
		}

		err = reviewCmd.Parse(commandArgs)
		if err != nil {
			if err == flag.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		}

		if *branchFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: For 'review', you must provide -b (branch).\n")
			reviewCmd.Usage()
			os.Exit(1)
		}

		if *repoFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: For 'review', you must provide -r (repository).\n")
			reviewCmd.Usage()
			os.Exit(1)
		}

		var prIDs []string
		if *prIDFlag != "" {
			prIDs = strings.Split(*prIDFlag, ",")
		}

		if err := app.HandleReview(ctx, cfg, defaultRunner, *branchFlag, prIDs, *repoFlag, *userFlag); err != nil {
			logger.Fatalf("Error handling review: %v", err)
		}

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
			os.Exit(1)
		}

		if (*pkgFlag != "" && *maintainerFlag != "") || (*pkgFlag == "" && *maintainerFlag == "") {
			fmt.Fprintf(os.Stderr, "Error: For 'bugowner', you must provide either -p (package) OR -m (maintainer), but not both.\n")
			bugownerCmd.Usage()
			os.Exit(1)
		}

		if *pkgFlag != "" {
			// Updated call to pass context and runner
			if err := app.HandleBugownerByPackage(ctx, cfg, defaultRunner, *pkgFlag); err != nil {
				logger.Fatalf("Error handling bugowner by package: %v", err)
			}
		} else { // *maintainerFlag != ""
			// Updated call to pass context and runner
			if err := app.HandlePackagesByMaintainer(ctx, cfg, defaultRunner, *maintainerFlag); err != nil {
				logger.Fatalf("Error handling packages by maintainer: %v", err)
			}
		}
	case "artifact":
		artifactCmd := flag.NewFlagSet("artifact", flag.ContinueOnError)
		projectFlag := artifactCmd.String("p", "", "Specify the project to list artifacts from (mandatory)")

		artifactCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s artifact:\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "  -p, --project <project>   List all artifacts for a specific project\n")
		}

		err = artifactCmd.Parse(commandArgs)
		if err != nil {
			if err == flag.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		}

		if *projectFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: a project must be specified using -p or --project.\n")
			artifactCmd.Usage()
			os.Exit(1)
		}

		if err := app.HandleArtifacts(ctx, cfg, defaultRunner, *projectFlag); err != nil {
			logger.Fatalf("Error handling artifacts: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s. Possible commands are:\n", command)
		for _, cmd := range validCommands {
			fmt.Printf("  %s\n", cmd)
		}
		os.Exit(1)
	}
}
