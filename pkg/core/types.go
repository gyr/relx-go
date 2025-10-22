package core

// PullRequest is a normalized structure representing a pull request result.
// Gitea will populate this via JSON unmarshaling.
type PullRequest struct {
	// ID is the unique identifier for the pull request.
	ID int `json:"id"`
	// Title is the title of the pull request.
	Title string `json:"title"`
	// State is the current state of the pull request (e.g., "open", "closed").
	State string `json:"state"`
	// URL is the HTML URL of the pull request.
	URL string `json:"html_url"`
}

// BuildStatus is a normalized structure for build results.
// OBS will populate this via XML unmarshaling.
type BuildStatus struct {
	// Project is the name of the OBS project.
	Project string `xml:"project,attr"`
	// Package is the name of the package.
	Package string `xml:"package,attr"`
	// Repository is the name of the repository.
	Repository string `xml:"repository,attr"`
	// Status is the build status of the package (e.g., "succeeded", "failed").
	Status string `xml:"code,attr"`
}