package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// VersionData contains all version-related data.
type VersionData struct {
	Version       string
	LatestVersion string
	GoVersion     string
	Platform      string
	HasUpdate     bool
	Error         error
}

// GitHubTag represent the data structure returned by GitHub API.
type GitHubTag struct {
	Name string `json:"name"`
}

// FetchLatestTag retrieves the latest GitHub tag for a repository.
func FetchLatestTag(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)

	if err != nil {
		return "", fmt.Errorf("error fetching tags: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var tags []GitHubTag

	if err := json.Unmarshal(body, &tags); err != nil {
		return "", fmt.Errorf("error decoding JSON: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found in GitHub repo")
	}

	return tags[0].Name, nil
}

// GetVersionInfo asynchronously fetches version metadata.
func GetVersionInfo(currentVersion, goVersion, platform string) (*VersionData, chan bool) {
	info := &VersionData{
		Version:   currentVersion,
		GoVersion: goVersion,
		Platform:  platform,
	}

	done := make(chan bool)

	go func() {
		defer close(done)

		latest, err := FetchLatestTag("MineHubs-Studios", "PreFlight")

		if err != nil {
			info.Error = err
			info.LatestVersion = "Unable to check"
			return
		}

		info.LatestVersion = latest
		info.HasUpdate = currentVersion != latest
	}()

	return info, done
}
