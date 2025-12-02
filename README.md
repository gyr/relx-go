# relx-go (Release eXtension for GO)

relx-go is a fast, dependency-minimal **Command-Line Interface (CLI) tool** written in Go for orchestrating release management tasks related to the openSUSE/SUSE ecosystem.

It serves as a unified wrapper for key external APIs and tools, such as the `osc` (OpenSUSE Commander) API for build status and the `git-obs` API for Gitea interactions.

The resulting binary, **`relx-go`**, is self-contained and highly portable.

## ✨ Features

*   **OBS Build Status:** Query the OBS API (`osc api`) for the build status of a specific package in a given project.

*   **Gitea Pull Requests:** Query the Gitea API (`git-obs api`) for a list of open pull requests in a given repository.

*   **Single Binary:** Zero runtime dependencies (beyond the `git-obs` and `osc` commands themselves).

*   **Modular and Testable:** The code is organized into a modular and testable architecture, with a clear separation of concerns between the command-line interface, application logic, and API clients.

*   **Well-tested:** The project has a comprehensive suite of unit tests to ensure the reliability of the code.

*   **Configurable:** Supports loading configuration from a Lua file, allowing customization of various settings like cache directories.

*   **Debug Logging:** Provides verbose output for troubleshooting and development purposes, controllable via configuration or command-line flag.

---

## ⚙️ Installation & Build

Since this is a CLI application, you can easily build it using the Go toolchain.

### Prerequisites

1.  [Go 1.21+](https://go.dev/doc/install)

2.  The external command-line tools (`osc` and `git-obs api`) must be installed and accessible in your system's PATH.

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

To ensure code quality and consistency, the following commands are used:

*   **Dependency Management:** `go mod tidy`
*   **Code Formatting:** `go fmt ./...`
*   **Static Analysis:** `go vet ./...`
*   **Build:** `go build ./...`
*   **Linting:** `golangci-lint run`

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

The following commands are recommended for a CI/CD pipeline to validate the code:

1.  **Install Dependencies:** `go get github.com/stretchr/testify/assert`
2.  **Tidy Modules:** `go mod tidy`
3.  **Run Tests:** `go test ./...`
4.  **Build for Production:** `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o relx-go cmd/relx-go/main.go`

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
cache_dir: "/path/to/your/custom/cache" # Customize the directory for cloning repositories
debug: true # Set to true to enable verbose debug logging
```

### Command-line Flags

Command-line flags provide a way to override or supplement configuration settings.

*   `-c`, `--config <path>`: Specify the path to a custom configuration file.
*   `-d`, `--debug`: Enable verbose debug logging. This flag overrides any `debug` setting in the configuration file.

## 🚀 Usage

The primary executable is `relx-go`. Commands are dispatched to the appropriate backend based on the subcommand used.

### 1. Check Pull Request Status (Gitea Backend)

Use the `pr` subcommand to list open pull requests for a given owner and repository.

| Argument | Description                  |
| -------- | ---------------------------- |
| `owner`  | The repository owner/organization. |
| `repo`   | The repository name.         |

```bash
./relx-go pr openSUSE osc
```

**Example Output:**

```
--- Open Pull Requests in openSUSE/osc ---
[101] Fix: Critical bug (State: open, URL: http://gitea/pr/101)
[102] Feature: New build step (State: open, URL: http://gitea/pr/102)
```

### 2. Check OBS Build Status (OBS Backend)

Use the `status` subcommand to check the build status for a specific package in an OBS project.

| Argument  | Description           |
| --------- | --------------------- |
| `project` | The OBS Project name. |
| `package` | The package name.     |

```bash
./relx-go status openSUSE:Factory osc
```

**Example Output:**

```
--- OBS Build Results for openSUSE:Factory/osc ---
Project: openSUSE:Factory, Package: osc, Repo: openSUSE_Tumbleweed, Status: succeeded
Project: openSUSE:Factory, Package: osc, Repo: openSUSE_Leap:15.5, Status: failed
```
