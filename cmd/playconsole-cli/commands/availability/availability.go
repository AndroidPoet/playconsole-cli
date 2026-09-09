package availability

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// AvailabilityCmd manages country targeting for releases
var AvailabilityCmd = &cobra.Command{
	Use:   "availability",
	Short: "Manage country availability for releases",
	Long: `View and update which countries your app is available in.

Country targeting is configured per release track. You can
restrict availability to specific countries or include all
countries with specific exclusions.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List country targeting for a track",
	RunE:  runList,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update country targeting for a track",
	RunE:  runUpdate,
}

var (
	track       string
	countries   string
	includeRest bool
)

func init() {
	listCmd.Flags().StringVar(&track, "track", "production", "release track")

	updateCmd.Flags().StringVar(&track, "track", "production", "release track")
	updateCmd.Flags().StringVar(&countries, "countries", "", "comma-separated country codes (e.g., US,GB,DE)")
	updateCmd.Flags().BoolVar(&includeRest, "include-rest", false, "include rest of world")
	updateCmd.Flags().Bool("confirm", false, "confirm destructive operation")
	cli.MustMarkFlagRequired(updateCmd, "countries")

	AvailabilityCmd.AddCommand(listCmd)
	AvailabilityCmd.AddCommand(updateCmd)
}

// CountryInfo represents country targeting information
type CountryInfo struct {
	Track          string   `json:"track"`
	Countries      []string `json:"countries,omitempty"`
	IncludeRest    bool     `json:"include_rest_of_world"`
	ReleaseVersion string   `json:"release_version,omitempty"`
}

// selectRelease picks the release whose country targeting is shown or
// updated: the in-progress rollout if any, else the completed release,
// else the first release on the track.
func selectRelease(releases []*androidpublisher.TrackRelease) *androidpublisher.TrackRelease {
	if len(releases) == 0 {
		return nil
	}
	for _, status := range []string{"inProgress", "completed"} {
		for _, r := range releases {
			if r.Status == status {
				return r
			}
		}
	}
	return releases[0]
}

func runList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	// Create edit to read track info
	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	trackResp, err := edit.Tracks().Get(
		client.GetPackageName(), edit.ID(), track,
	).Context(edit.Context()).Do()
	if err != nil {
		return fmt.Errorf("failed to get track '%s': %w", track, err)
	}

	info := CountryInfo{
		Track: track,
	}

	if release := selectRelease(trackResp.Releases); release != nil {
		if release.CountryTargeting != nil {
			info.Countries = release.CountryTargeting.Countries
			info.IncludeRest = release.CountryTargeting.IncludeRestOfWorld
		}
		info.ReleaseVersion = release.Name
		if info.ReleaseVersion == "" && len(release.VersionCodes) > 0 {
			info.ReleaseVersion = fmt.Sprintf("%d", release.VersionCodes[0])
		}
	}

	return output.Print(info)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}

	countryCodes := strings.Split(countries, ",")
	for i := range countryCodes {
		countryCodes[i] = strings.TrimSpace(strings.ToUpper(countryCodes[i]))
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update track '%s' with countries: %s (include rest: %v)",
			track, strings.Join(countryCodes, ", "), includeRest)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	// Create edit
	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	// Get current track
	trackResp, err := edit.Tracks().Get(
		client.GetPackageName(), edit.ID(), track,
	).Context(edit.Context()).Do()
	if err != nil {
		return fmt.Errorf("failed to get track '%s': %w", track, err)
	}

	release := selectRelease(trackResp.Releases)
	if release == nil {
		return fmt.Errorf("no releases found on track '%s'", track)
	}

	// Update country targeting on the selected release only
	release.CountryTargeting = &androidpublisher.CountryTargeting{
		Countries:          countryCodes,
		IncludeRestOfWorld: includeRest,
	}

	// Update track
	_, err = edit.Tracks().Update(
		client.GetPackageName(), edit.ID(), track, trackResp,
	).Context(edit.Context()).Do()
	if err != nil {
		return fmt.Errorf("failed to update track: %w", err)
	}

	// Commit
	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Country targeting updated for track '%s'", track)
	return nil
}
