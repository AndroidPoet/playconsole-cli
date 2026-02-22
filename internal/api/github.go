package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	githubAPIURL = "https://api.github.com/repos/AndroidPoet/playconsole-cli/releases"
)

// GitHubClient handles GitHub API requests
type GitHubClient struct {
	httpClient *http.Client
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	PublishedAt string        `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a release asset
type GitHubAsset struct {
	Name          string `json:"name"`
	DownloadCount int64  `json:"download_count"`
	DownloadURL   string `json:"browser_download_url"`
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetReleases fetches all releases from GitHub
func (c *GitHubClient) GetReleases() ([]GitHubRelease, error) {
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "playconsole-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return releases, nil
}

// GetTotalDownloads returns the total download count across all releases
func (c *GitHubClient) GetTotalDownloads() (int64, error) {
	releases, err := c.GetReleases()
	if err != nil {
		return 0, err
	}

	var total int64
	for _, release := range releases {
		for _, asset := range release.Assets {
			total += asset.DownloadCount
		}
	}

	return total, nil
}

// GetDownloadsByRelease returns download counts grouped by release
func (c *GitHubClient) GetDownloadsByRelease() ([]ReleaseDownloads, error) {
	releases, err := c.GetReleases()
	if err != nil {
		return nil, err
	}

	result := make([]ReleaseDownloads, 0, len(releases))
	for _, release := range releases {
		var downloads int64
		for _, asset := range release.Assets {
			downloads += asset.DownloadCount
		}

		// Parse date
		date := release.PublishedAt
		if t, err := time.Parse(time.RFC3339, release.PublishedAt); err == nil {
			date = t.Format("2006-01-02")
		}

		result = append(result, ReleaseDownloads{
			Tag:       release.TagName,
			Date:      date,
			Downloads: downloads,
		})
	}

	return result, nil
}

// GetDownloadsByPlatform returns download counts grouped by platform
func (c *GitHubClient) GetDownloadsByPlatform() ([]PlatformDownloads, int64, error) {
	releases, err := c.GetReleases()
	if err != nil {
		return nil, 0, err
	}

	platformCounts := make(map[string]int64)
	var total int64

	for _, release := range releases {
		for _, asset := range release.Assets {
			platform := extractPlatform(asset.Name)
			platformCounts[platform] += asset.DownloadCount
			total += asset.DownloadCount
		}
	}

	result := make([]PlatformDownloads, 0, len(platformCounts))
	for platform, downloads := range platformCounts {
		percent := float64(0)
		if total > 0 {
			percent = float64(downloads) / float64(total) * 100
		}
		result = append(result, PlatformDownloads{
			Platform:  platform,
			Downloads: downloads,
			Percent:   fmt.Sprintf("%.1f%%", percent),
		})
	}

	return result, total, nil
}

// ReleaseDownloads represents download stats for a release
type ReleaseDownloads struct {
	Tag       string `json:"tag"`
	Date      string `json:"date"`
	Downloads int64  `json:"downloads"`
}

// PlatformDownloads represents download stats for a platform
type PlatformDownloads struct {
	Platform  string `json:"platform"`
	Downloads int64  `json:"downloads"`
	Percent   string `json:"percent"`
}

// extractPlatform extracts the platform from an asset name
func extractPlatform(name string) string {
	name = strings.ToLower(name)

	switch {
	case strings.Contains(name, "darwin_arm64"):
		return "darwin_arm64"
	case strings.Contains(name, "darwin_amd64"):
		return "darwin_amd64"
	case strings.Contains(name, "linux_arm64"):
		return "linux_arm64"
	case strings.Contains(name, "linux_amd64"):
		return "linux_amd64"
	case strings.Contains(name, "windows_arm64"):
		return "windows_arm64"
	case strings.Contains(name, "windows_amd64"):
		return "windows_amd64"
	case strings.Contains(name, "checksum"):
		return "checksums"
	default:
		return "other"
	}
}
