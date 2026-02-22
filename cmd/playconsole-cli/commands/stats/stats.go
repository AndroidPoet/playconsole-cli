package stats

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// StatsCmd is the root command for CLI statistics
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View CLI download statistics",
	Long: `View download statistics for the playconsole-cli tool itself.

This tracks downloads from GitHub releases and shows distribution
across platforms and versions.`,
}

var downloadsCmd = &cobra.Command{
	Use:   "downloads",
	Short: "View download counts",
	Long: `View download statistics from GitHub releases.

Shows total downloads, with optional breakdowns by release or platform.`,
	RunE: runDownloads,
}

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "Show download sources and tracking status",
	RunE:  runSources,
}

var (
	byRelease  bool
	byPlatform bool
)

func init() {
	downloadsCmd.Flags().BoolVar(&byRelease, "by-release", false, "show breakdown by release version")
	downloadsCmd.Flags().BoolVar(&byPlatform, "by-platform", false, "show breakdown by platform")

	StatsCmd.AddCommand(downloadsCmd)
	StatsCmd.AddCommand(sourcesCmd)
}

// DownloadStats represents overall download statistics
type DownloadStats struct {
	Total       int64                   `json:"total"`
	ByRelease   []api.ReleaseDownloads  `json:"by_release,omitempty"`
	ByPlatform  []api.PlatformDownloads `json:"by_platform,omitempty"`
	LastUpdated string                  `json:"last_updated"`
}

// DownloadSource represents a download source
type DownloadSource struct {
	Name        string `json:"name"`
	Tracked     bool   `json:"tracked"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
}

func runDownloads(cmd *cobra.Command, args []string) error {
	client := api.NewGitHubClient()

	if byPlatform {
		platforms, _, err := client.GetDownloadsByPlatform()
		if err != nil {
			return fmt.Errorf("failed to fetch platform stats: %w", err)
		}
		return output.Print(platforms)
	}

	if byRelease {
		releases, err := client.GetDownloadsByRelease()
		if err != nil {
			return fmt.Errorf("failed to fetch release stats: %w", err)
		}
		return output.Print(releases)
	}

	// Default: show summary
	total, err := client.GetTotalDownloads()
	if err != nil {
		return fmt.Errorf("failed to fetch download stats: %w", err)
	}

	stats := DownloadStats{
		Total:       total,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}

	return output.Print(stats)
}

func runSources(cmd *cobra.Command, args []string) error {
	sources := []DownloadSource{
		{
			Name:        "GitHub Releases",
			Tracked:     true,
			Description: "Direct downloads from releases page",
			Command:     "curl -LO https://github.com/.../releases/...",
		},
		{
			Name:        "Install Script",
			Tracked:     true,
			Description: "curl | bash (downloads from GitHub)",
			Command:     "curl -fsSL .../install.sh | bash",
		},
		{
			Name:        "Homebrew",
			Tracked:     false,
			Description: "brew install (stats not publicly available)",
			Command:     "brew install playconsole-cli",
		},
		{
			Name:        "Go Install",
			Tracked:     false,
			Description: "go install (no tracking available)",
			Command:     "go install github.com/AndroidPoet/playconsole-cli@latest",
		},
	}

	return output.Print(sources)
}
