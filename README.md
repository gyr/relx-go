# relx-go (Release eXtension for GO)

relx-go is a fast, dependency-minimal **Command-Line Interface (CLI) tool** written in Go for orchestrating release management tasks related to the openSUSE/SUSE ecosystem.

It serves as a unified wrapper for key external APIs and tools, such as the `osc` (OpenSUSE Commander) API for build status and the `git-obs` API for Gitea interactions.

The resulting binary, **`relx-go`**, is self-contained and highly portable.

## ✨ Features

*   **OBS Artifacts:** Fetch and filter binary artifacts for a specific package in a given OBS project.

*   **Gitea Pull Requests:** Interactively review open pull requests from Gitea. This includes listing PRs based on reviewer, branch, and repository, viewing content and diffs using `delta`, and taking actions like approving (placeholder), skipping, or exiting the review process.

*   **Single Binary:** Zero runtime dependencies (beyond the `git-obs` and `osc` commands themselves).

*   **Modular and Testable:** The code is organized into a modular and testable architecture, with a clear separation of concerns between the command-line interface, application logic, and API clients.

*   **Well-tested:** The project has a comprehensive suite of unit tests to ensure the reliability of the code.

*   **Configurable:** Supports loading configuration from a YAML file, allowing customization of various settings like cache directories and API endpoints.

*   **Debug Logging:** Provides verbose output for troubleshooting and development purposes, controllable via configuration or command-line flag.

---

## ⚙️ Installation & Build

Since this is a CLI application, you can easily build it using the Go toolchain.

### Prerequisites

1.  [Go 1.21+](https://go.dev/doc/install)

2.  The external command-line tools (`osc`, `git-obs api`, and `delta`) must be installed and accessible in your system's PATH.

### Building the Executable

There are two main ways to build the executable, depending on your needs.

#### For Local Development

For day-to-day development on your local machine, use the standard `go build` command. This is the quickest and easiest way to get a running executable for your current operating system and architecture.

```bash
# Build the executable for your current OS/Arch
go build -o relx-go cmd/relx-go/main.go
```

Alternatively, you can use `go run` to compile and execute the application in a single step without creating a persistent binary. This is useful for quickly testing code changes.

```bash
# Run the application directly
go run cmd/relx-go/main.go [args]
```
For example, to run the `pr` subcommand:
```bash
go run cmd/relx-go/main.go pr openSUSE osc
```
```

#### For Production and CI/CD (Best Practice)

For creating a distributable binary for a production environment, a CI/CD pipeline, or for sharing with others, it is a best practice to create a statically-linked binary for a specific target. This ensures portability and reproducibility.

The following command creates a statically-linked binary for 64-bit Linux:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o relx-go cmd/relx-go/main.go
```

*   **`CGO_ENABLED=0`**: Disables Cgo, which creates a statically linked binary with no external dependencies. This makes the executable highly portable.
*   **`GOOS=linux`**: Sets the target Operating System to Linux.
*   **`GOARCH=amd64`**: Sets the target CPU Architecture to amd64 (64-bit).

### Code Quality and Formatting

Before committing your changes, it is a best practice to run the following commands locally. These tools help automatically format your code, manage dependencies, and run tests, ensuring the code you commit is clean and correct.

*   **Format Code (`go fmt ./...`):** This command automatically reformats all Go source files in the project to follow the official Go style guidelines. It ensures consistency across the entire codebase.

*   **Tidy Dependencies (`go mod tidy`):** This command analyzes your source code and updates your `go.mod` and `go.sum` files. It adds any new dependencies that are required and removes any that are no longer used. It is essential for keeping your project's dependency list accurate.

*   **Run Linter (`golangci-lint run`):** This runs a powerful linter that checks for a wide variety of programming mistakes, styling issues, and potential bugs. It implicitly runs `go vet` and many other checks, making a separate `go vet` step redundant.

*   **Build Application (`go build -v ./...`):** This command compiles the entire application. It's a great final check to ensure all the code is syntactically correct and the project builds successfully as a whole.

*   **Run Tests (`go test -v -race ./...`):** This runs all the unit tests in the project. The `-race` flag enables Go's race detector, which is invaluable for finding and debugging concurrency problems in your code.

### Troubleshooting Build Issues

If you encounter unexpected build errors, particularly related to dependency resolution or stale module information (e.g., "MissingFieldOrMethod" warnings after updating code), it's often helpful to clear your Go module cache.

```bash
go clean -modcache
```

This command removes the entire Go module cache, forcing Go to re-download all dependencies on the next build. This ensures that you are working with the freshest versions of your modules, as specified in your `go.mod` file.

## ✅ Testing

relx-go uses Go's built-in testing package. All tests mock the external command execution to ensure fast and reliable unit testing without relying on the network or installed external binaries.

To run all unit tests in the entire project, execute the following command from the project root:

```bash
go test ./...
```

To run tests with verbose output:

```bash
go test -v ./...
```

## CI Validation

The Continuous Integration (CI) pipeline automatically validates all code that is pushed to the repository. Its purpose is not to *fix* the code, but to *verify* that the code is correct and adheres to the project's standards. This acts as a safety net. The following steps are performed:

1.  **Verify Modules (`go mod verify`):** This command checks that the dependencies recorded in `go.sum` are consistent and haven't been tampered with. Unlike `go mod tidy`, it doesn't change any files; it just verifies the state of the committed dependency files.

2.  **Linting:** This step runs `golangci-lint` to check for code quality issues. It ensures that any committed code has already passed the linter checks. The CI will fail if the linter finds any problems.

3.  **Build (`go build -v ./...`):** This command compiles the entire application to ensure that the code is syntactically correct and all dependencies can be resolved.

4.  **Run Tests (`go test -v -race ./...`):** This step runs the same test command that is used locally, including the race detector. It confirms that all tests pass on a clean environment, ensuring the changes haven't introduced any regressions.

## ⚙️ Configuration

relx-go supports loading configuration from a YAML file. This allows you to customize settings such as the cache directory used for cloning Git repositories and enable debug logging.

### Configuration File Search Order

relx-go searches for a configuration file in the following order, using the first one found:

1.  **Command-line flag:** Specified using `--config <path>` or `-c <path>`.
2.  **Environment variable:** The path specified in the `RELX_GO_CONFIG_FILE` environment variable.
3.  **User-specific path:** `~/.config/relx-go/config.yaml`
4.  **System-wide path:** `/etc/relx-go/config.yaml`

If no configuration file is found, relx-go will proceed with default settings.

### Example `config.yaml`

```yaml
cache_dir: "~/.cache/relx-go"
debug: true
repo_url: "https://example.com/user/repo.git"
repo_branch: "main"
operation_timeout_seconds: 300
obs_api_url: "https://api.suse.de"

# Filter patterns for OBS packages
package_filter_patterns:
  - "000productcompose:sles_*"
  - "SLES_transactional:*"

# Filter patterns for binary artifacts
binary_filter_patterns:
  - "*.iso"
  - "*.qcow2"
```

### Command-line Flags

Command-line flags provide a way to override or supplement configuration settings.

*   `-c`, `--config <path>`: Specify the path to a custom configuration file.
*   `-d`, `--debug`: Enable verbose debug logging. This flag overrides any `debug` setting in the configuration file.

## 🚀 Usage

The primary executable is `relx-go`. Commands are dispatched to the appropriate backend based on the subcommand used.

---
### 1. Review Pull Requests (Gitea Backend)

Use the `review` subcommand to interactively review open pull requests for a given branch and repository.

The workflow is as follows:
1.  The command lists all open pull requests for the specified reviewer, branch, and repository.
2.  You will be prompted if you wish to proceed with reviewing these pull requests.
3.  If you confirm, each pull request's timeline and patch (diff) will be displayed using `delta` (allowing you to scroll and inspect changes).
4.  After reviewing each PR, you will be prompted to 'approve' (placeholder for future implementation), 'skip' (move to the next PR), or 'exit' (terminate the review process).

| Flag      | Description                  |
| --------- | ---------------------------- |
| `-b`      | The branch to review.         |
| `-r`      | The repository to review.     |

**Note:** The `pr_reviewer` configuration must be set in your `config.yaml` file for this subcommand to work.

```bash
./relx-go review -b master -r osc
```

**Example Output (interactive flow):**

```
--- Open Pull Requests for Review ---
PR ID: 499
PR ID: 496
Do you want to review these pull requests? (y/n): y
(Delta will now display PR 499 content and diff, interact with Delta)
Approve, skip, or exit? (a/s/e): s
Skipping PR 499.
(Delta will now display PR 496 content and diff, interact with Delta)
Approve, skip, or exit? (a/s/e): a
PR 496 approved (future implementation).
```

### 2. List OBS Artifacts (OBS Backend)

Use the `artifact` subcommand to list binary artifacts for a specific project in OBS.

| Flag      | Description                |
| --------- | -------------------------- |
| `-p`      | The OBS Project name.      |

```bash
./relx-go artifact -p SUSE:SLFO:Product:SLES:16.1
```

**Example Output:**

```
Artifacts for project 'SUSE:SLFO:Product:SLES:16.1':
SLE-16.1-Installer-DVD-x86_64-Build1.1.iso
SLE-16.1-Installer-DVD-x86_64-Build1.1.qcow2
```
