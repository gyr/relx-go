package gitutils

import (
	"errors"
	"os"
	"os/user"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

func TestCloneRepo_Success(t *testing.T) {
	originalUserCurrent := userCurrent
	originalOsMkdirAll := osMkdirAll
	originalExecCommand := execCommand
	defer func() {
		userCurrent = originalUserCurrent
		osMkdirAll = originalOsMkdirAll
		execCommand = originalExecCommand
	}()

	userCurrent = func() (*user.User, error) {
		return &user.User{HomeDir: "/home/testuser"}, nil
	}

	osMkdirAll = func(path string, perm os.FileMode) error {
		return nil
	}

	execCommand = func(name string, args ...string) ([]byte, error) {
		return []byte("success"), nil
	}

	mockCfg := &config.Config{
		RepoURL:  "https://example.com/test.git",
		RepoName: "test",
		CacheDir: "/home/testuser/.cache/relx-go",
		Logger:   logging.NewLogger(logging.LevelDebug),
	}

	path, err := CloneRepo(mockCfg)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	expectedPath := mockCfg.CacheDir + "/" + mockCfg.RepoName
	if path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, path)
	}
}

func TestCloneRepo_ExecFailure(t *testing.T) {
	originalUserCurrent := userCurrent
	originalOsMkdirAll := osMkdirAll
	originalExecCommand := execCommand
	defer func() {
		userCurrent = originalUserCurrent
		osMkdirAll = originalOsMkdirAll
		execCommand = originalExecCommand
	}()

	userCurrent = func() (*user.User, error) {
		return &user.User{HomeDir: "/home/testuser"}, nil
	}

	osMkdirAll = func(path string, perm os.FileMode) error {
		return nil
	}

	mockError := errors.New("clone failed")
	execCommand = func(name string, args ...string) ([]byte, error) {
		return []byte("error"), mockError
	}

	mockCfg := &config.Config{
		RepoURL:  "https://example.com/test.git",
		RepoName: "test",
		CacheDir: "/home/testuser/.cache/relx-go",
		Logger:   logging.NewLogger(logging.LevelDebug),
	}

	_, err := CloneRepo(mockCfg)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), mockError.Error()) {
		t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
	}
}

func TestCloneRepo_EmptyCacheDir(t *testing.T) {
	originalUserCurrent := userCurrent
	originalOsMkdirAll := osMkdirAll
	originalExecCommand := execCommand
	defer func() {
		userCurrent = originalUserCurrent
		osMkdirAll = originalOsMkdirAll
		execCommand = originalExecCommand
	}()

	userCurrent = func() (*user.User, error) {
		return &user.User{HomeDir: "/home/testuser"}, nil
	}

	osMkdirAll = func(path string, perm os.FileMode) error {
		return nil
	}

	execCommand = func(name string, args ...string) ([]byte, error) {
		return []byte("success"), nil
	}

	mockCfg := &config.Config{
		RepoURL:  "https://example.com/test.git",
		RepoName: "test",
		CacheDir: "", // Empty cacheDir to test default behavior
		Logger:   logging.NewLogger(logging.LevelDebug),
	}

	path, err := CloneRepo(mockCfg)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	expectedPath := "/home/testuser/.cache/relx-go/test"
	if path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, path)
	}
}
