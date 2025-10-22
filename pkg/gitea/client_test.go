package gitea

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestGetPullRequests_Success(t *testing.T) {
	// Define the JSON response we expect from 'git-obs api'
	mockJSON := `{
		"issues": [
			{"id": 101, "title": "Fix: Critical bug", "state": "open", "html_url": "http://gitea/pr/101"},
			{"id": 102, "title": "Feature: New build step", "state": "open", "html_url": "http://gitea/pr/102"}
		]
	}`

	// Temporarily replace the external command executor function
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(name string, args ...string) ([]byte, error) {
		return []byte(mockJSON), nil
	}

	client := NewClient("/tmp/cache") // Add dummy cacheDir
	prs, err := client.GetPullRequests("testowner", "testrepo")

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if len(prs) != 2 {
		t.Fatalf("Expected 2 pull requests, got %d", len(prs))
	}

	expectedTitle := "Fix: Critical bug"
	if prs[0].Title != expectedTitle {
		t.Errorf("PR[0] Title: got %q, want %q", prs[0].Title, expectedTitle)
	}
}

func TestGetPullRequests_ExecutionFailure(t *testing.T) {
	// Temporarily replace the external command executor function
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	mockError := errors.New("command failed")
	execCommand = func(name string, args ...string) ([]byte, error) {
		return nil, mockError
	}

	client := NewClient("/tmp/cache") // Add dummy cacheDir
	_, err := client.GetPullRequests("badowner", "badrepo")

	if err == nil {
		t.Fatal("Expected an error due to command failure, but got nil")
	}

	if !strings.Contains(err.Error(), mockError.Error()) {
		t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
	}
}

func TestBuildGiteaURL(t *testing.T) {
	tests := []struct {
		owner    string
		expected string
	}{
		{"my-owner", "/repos/issues/search?limit=50&owner=my-owner&state=open&type=pulls"},
		{"another_owner", "/repos/issues/search?limit=50&owner=another_owner&state=open&type=pulls"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("owner=%s", tt.owner), func(t *testing.T) {
			actual, err := buildGiteaURL(tt.owner)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != tt.expected {
				t.Errorf("expected URL %q, got %q", tt.expected, actual)
			}
		})
	}
}
