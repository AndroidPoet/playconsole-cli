package diff

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// DiffCmd compares edit state vs live version
var DiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between draft edit and live app",
	Long: `Compare the current draft edit state against the live version.

This helps you review pending changes before committing an edit.
If no edit exists, shows the current live state as a summary.`,
	RunE: runDiff,
}

var (
	editID  string
	section string
)

func init() {
	DiffCmd.Flags().StringVar(&editID, "edit-id", "", "existing edit ID to compare (optional)")
	DiffCmd.Flags().StringVar(&section, "section", "all", "section to diff: listings, tracks, all")
}

// DiffResult represents the diff output
type DiffResult struct {
	Section  string      `json:"section"`
	Status   string      `json:"status"`
	Current  interface{} `json:"current,omitempty"`
}

// TrackSummary represents a track's current state
type TrackSummary struct {
	Track       string `json:"track"`
	Status      string `json:"status,omitempty"`
	VersionCode int64  `json:"version_code,omitempty"`
	UserFraction float64 `json:"user_fraction,omitempty"`
}

// ListingSummary represents a listing's current state
type ListingSummary struct {
	Language string `json:"language"`
	Title    string `json:"title,omitempty"`
}

func runDiff(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	// Get or create an edit to read current state
	var edit *api.Edit
	if editID != "" {
		edit, err = client.GetEdit(editID)
	} else {
		edit, err = client.CreateEdit()
	}
	if err != nil {
		return err
	}
	defer edit.Close()
	defer edit.Delete()

	results := make([]DiffResult, 0)

	// Listings diff
	if section == "all" || section == "listings" {
		listingsResp, err := edit.Listings().List(
			client.GetPackageName(), edit.ID(),
		).Context(edit.Context()).Do()
		if err != nil {
			results = append(results, DiffResult{
				Section: "listings",
				Status:  fmt.Sprintf("error: %v", err),
			})
		} else {
			summaries := make([]ListingSummary, 0, len(listingsResp.Listings))
			for _, l := range listingsResp.Listings {
				summaries = append(summaries, ListingSummary{
					Language: l.Language,
					Title:    l.Title,
				})
			}
			results = append(results, DiffResult{
				Section: "listings",
				Status:  fmt.Sprintf("%d locales", len(listingsResp.Listings)),
				Current: summaries,
			})
		}
	}

	// Tracks diff
	if section == "all" || section == "tracks" {
		tracksResp, err := edit.Tracks().List(
			client.GetPackageName(), edit.ID(),
		).Context(edit.Context()).Do()
		if err != nil {
			results = append(results, DiffResult{
				Section: "tracks",
				Status:  fmt.Sprintf("error: %v", err),
			})
		} else {
			summaries := make([]TrackSummary, 0, len(tracksResp.Tracks))
			for _, t := range tracksResp.Tracks {
				summary := TrackSummary{
					Track: t.Track,
				}
				if len(t.Releases) > 0 {
					latest := t.Releases[0]
					summary.Status = latest.Status
					if len(latest.VersionCodes) > 0 {
						summary.VersionCode = latest.VersionCodes[0]
					}
					summary.UserFraction = latest.UserFraction
				}
				summaries = append(summaries, summary)
			}
			results = append(results, DiffResult{
				Section: "tracks",
				Status:  fmt.Sprintf("%d tracks", len(tracksResp.Tracks)),
				Current: summaries,
			})
		}
	}

	return output.Print(results)
}
