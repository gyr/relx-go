package gitutils

import (
	"testing"
)

func TestDeriveRepoName(t *testing.T) {
	testCases := []struct {
		name     string
		repoURL  string
		expected string
		err      bool
	}{
		{"Standard HTTPS", "https://example.com/user/repo.git", "repo", false},
		{"HTTPS without .git", "https://example.com/user/repo", "repo", false},
		{"Standard SSH", "git@example.com:user/repo.git", "repo", false},
		{"SSH without .git", "git@example.com:user/repo", "repo", false},
		{"Generic SSH", "gitea@example.com:user/another-repo.git", "another-repo", false},
		{"GitLab SSH with slash path", "gitlab@my-instance.com/user/repo.git", "repo", false},
		{"No name", "https://example.com/", "", true},
		{"Invalid URL", "https://", "", true},
		{"Empty URL", "", "", true},
		{"Just .git", ".git", "", true},
		{"Just a dot", ".", "", true},
		{"Colon only", ":", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoName, err := deriveRepoName(tc.repoURL)
			if (err != nil) != tc.err {
				t.Fatalf("Expected error: %v, got: %v", tc.err, err)
			}
			if repoName != tc.expected {
				t.Errorf("Expected repo name: %q, got: %q", tc.expected, repoName)
			}
		})
	}
}
