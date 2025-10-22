package gitutils

import (
	"errors"
	"os"
	"os/user"
	"strings"
	"testing"
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

	cacheDir := "/home/testuser/.cache/grxs"
	path, err := CloneRepo("https://example.com/test.git", cacheDir)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	expectedPath := cacheDir + "/test.git"
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

	cacheDir := "/home/testuser/.cache/grxs"
	_, err := CloneRepo("https://example.com/test.git", cacheDir)
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

	path, err := CloneRepo("https://example.com/test.git", "") // Empty cacheDir
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	expectedPath := "/home/testuser/.cache/grxs/test.git"
	if path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, path)
	}
}