package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// VersionData CONTAINS ALL VERSION-RELATED DATA.
type VersionData struct {
	Version       string
	LatestVersion string
	GoVersion     string
	Platform      string
	HasUpdate     bool
	Error         error
}

// GitHubTag REPRESENT THE DATA STRUCTURE RETURNED BY GitHub API.
type GitHubTag struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

// FetchLatestTag GETS THE LATEST TAG AND ITS PUBLISHING DATE FROM GitHub API.
func FetchLatestTag(repoOwner, repoName string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", repoOwner, repoName)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)

	if err != nil {
		return "", fmt.Errorf("error fetching tags: %w", err)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			_ = err // EXPLICITLY DISCARD THE ERROR TO SILENCE SA90003.
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var tags []GitHubTag

	if err := json.Unmarshal(body, &tags); err != nil {
		return "", fmt.Errorf("error parsing JSON: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found")
	}

	return tags[0].Name, nil
}

// GetVersionInfo COLLECTS ALL VERSION DATA.
func GetVersionInfo(currentVersion, goVersion, platform string) (*VersionData, chan bool) {
	info := &VersionData{
		Version:       currentVersion,
		LatestVersion: "",
		GoVersion:     goVersion,
		Platform:      platform,
	}

	done := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(1)

	// FETCH LATEST VERSION.
	go func() {
		latestVersion, err := FetchLatestTag("MineHubs-Studios", "PreFlight")

		if err != nil {
			info.Error = err
			info.LatestVersion = "Unable to check"
		} else {
			info.LatestVersion = latestVersion
		}

		close(done)
	}()

	return info, done
}
