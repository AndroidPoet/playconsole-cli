package tracks

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var TracksCmd = &cobra.Command{
	Use:   "tracks",
	Short: "Manage release tracks",
	Long: `Manage release tracks (internal, alpha, beta, production).

Tracks control how your app is distributed to users. Each track can have
different release configurations and rollout percentages.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracks",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get track details",
	RunE:  runGet,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a track",
	Long: `Update a track with a new release.

This command allows you to set version codes, rollout percentages,
release notes, and release status.`,
	RunE: runUpdate,
}

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote a release to another track",
	RunE:  runPromote,
}

var haltCmd = &cobra.Command{
	Use:   "halt",
	Short: "Halt a staged rollout",
	RunE:  runHalt,
}

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Complete a staged rollout (100%)",
	RunE:  runComplete,
}

var (
	trackName         string
	versionCode       int64
	versionCodes      []int64
	rolloutPercentage float64
	releaseNotes      string
	releaseNotesLang  string
	status            string
	fromTrack         string
	toTrack           string
)

func init() {
	// List has no additional flags

	// Get flags
	getCmd.Flags().StringVar(&trackName, "track", "", "track name (internal, alpha, beta, production)")
	getCmd.MarkFlagRequired("track")

	// Update flags
	updateCmd.Flags().StringVar(&trackName, "track", "", "track name")
	updateCmd.Flags().Int64Var(&versionCode, "version-code", 0, "version code to release")
	updateCmd.Flags().Int64SliceVar(&versionCodes, "version-codes", nil, "multiple version codes")
	updateCmd.Flags().Float64Var(&rolloutPercentage, "rollout-percentage", 100, "rollout percentage (0-100)")
	updateCmd.Flags().StringVar(&releaseNotes, "release-notes", "", "release notes text")
	updateCmd.Flags().StringVar(&releaseNotesLang, "release-notes-lang", "en-US", "release notes language")
	updateCmd.Flags().StringVar(&status, "status", "completed", "release status (draft, inProgress, halted, completed)")
	updateCmd.MarkFlagRequired("track")

	// Promote flags
	promoteCmd.Flags().StringVar(&fromTrack, "from", "", "source track")
	promoteCmd.Flags().StringVar(&toTrack, "to", "", "destination track")
	promoteCmd.Flags().Int64Var(&versionCode, "version-code", 0, "specific version code to promote (optional)")
	promoteCmd.Flags().Float64Var(&rolloutPercentage, "rollout-percentage", 100, "rollout percentage")
	promoteCmd.MarkFlagRequired("from")
	promoteCmd.MarkFlagRequired("to")

	// Halt flags
	haltCmd.Flags().StringVar(&trackName, "track", "", "track name")
	haltCmd.MarkFlagRequired("track")

	// Complete flags
	completeCmd.Flags().StringVar(&trackName, "track", "", "track name")
	completeCmd.MarkFlagRequired("track")

	TracksCmd.AddCommand(listCmd)
	TracksCmd.AddCommand(getCmd)
	TracksCmd.AddCommand(updateCmd)
	TracksCmd.AddCommand(promoteCmd)
	TracksCmd.AddCommand(haltCmd)
	TracksCmd.AddCommand(completeCmd)
}

// TrackInfo represents track information for output
type TrackInfo struct {
	Track        string   `json:"track"`
	VersionCodes []int64  `json:"version_codes,omitempty"`
	Status       string   `json:"status,omitempty"`
	Rollout      float64  `json:"rollout,omitempty"`
	ReleaseCount int      `json:"release_count"`
}

func runList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()
	defer edit.Delete()

	tracks, err := edit.Tracks().List(client.GetPackageName(), edit.ID()).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	result := make([]TrackInfo, 0, len(tracks.Tracks))
	for _, t := range tracks.Tracks {
		info := TrackInfo{
			Track:        t.Track,
			ReleaseCount: len(t.Releases),
		}

		// Get latest release info
		if len(t.Releases) > 0 {
			latest := t.Releases[0]
			info.VersionCodes = latest.VersionCodes
			info.Status = latest.Status

			if latest.UserFraction > 0 {
				info.Rollout = latest.UserFraction * 100
			} else if latest.Status == "completed" {
				info.Rollout = 100
			}
		}

		result = append(result, info)
	}

	return output.Print(result)
}

func runGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()
	defer edit.Delete()

	track, err := edit.Tracks().Get(client.GetPackageName(), edit.ID(), trackName).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	return output.Print(track)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	// Validate version codes
	codes := versionCodes
	if versionCode > 0 {
		codes = append(codes, versionCode)
	}
	if len(codes) == 0 {
		return fmt.Errorf("at least one version code is required (--version-code or --version-codes)")
	}

	// Validate rollout
	if rolloutPercentage < 0 || rolloutPercentage > 100 {
		return fmt.Errorf("rollout percentage must be between 0 and 100")
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	// Build release
	release := &androidpublisher.TrackRelease{
		VersionCodes: codes,
		Status:       status,
	}

	// Set user fraction for staged rollouts
	if rolloutPercentage < 100 && status == "inProgress" {
		release.UserFraction = rolloutPercentage / 100
	}

	// Add release notes if provided
	if releaseNotes != "" {
		release.ReleaseNotes = []*androidpublisher.LocalizedText{
			{
				Language: releaseNotesLang,
				Text:     releaseNotes,
			},
		}
	}

	// Build track update
	trackUpdate := &androidpublisher.Track{
		Track:    trackName,
		Releases: []*androidpublisher.TrackRelease{release},
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update track '%s' with version codes %v", trackName, codes)
		return output.Print(trackUpdate)
	}

	// Update track
	updated, err := edit.Tracks().Update(client.GetPackageName(), edit.ID(), trackName, trackUpdate).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	// Commit the edit
	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Track '%s' updated successfully", trackName)
	return output.Print(updated)
}

func runPromote(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	ctx := edit.Context()

	// Get source track
	sourceTrack, err := edit.Tracks().Get(client.GetPackageName(), edit.ID(), fromTrack).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get source track '%s': %w", fromTrack, err)
	}

	// Find version codes to promote
	var codesToPromote []int64
	if versionCode > 0 {
		codesToPromote = []int64{versionCode}
	} else if len(sourceTrack.Releases) > 0 {
		// Use latest release's version codes
		codesToPromote = sourceTrack.Releases[0].VersionCodes
	}

	if len(codesToPromote) == 0 {
		return fmt.Errorf("no version codes found in track '%s'", fromTrack)
	}

	// Build release for destination
	release := &androidpublisher.TrackRelease{
		VersionCodes: codesToPromote,
		Status:       "completed",
	}

	if rolloutPercentage < 100 {
		release.Status = "inProgress"
		release.UserFraction = rolloutPercentage / 100
	}

	destTrack := &androidpublisher.Track{
		Track:    toTrack,
		Releases: []*androidpublisher.TrackRelease{release},
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would promote version codes %v from '%s' to '%s'", codesToPromote, fromTrack, toTrack)
		return output.Print(destTrack)
	}

	// Update destination track
	updated, err := edit.Tracks().Update(client.GetPackageName(), edit.ID(), toTrack, destTrack).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Commit
	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Promoted version codes %v from '%s' to '%s'", codesToPromote, fromTrack, toTrack)
	return output.Print(updated)
}

func runHalt(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	ctx := edit.Context()

	// Get current track
	track, err := edit.Tracks().Get(client.GetPackageName(), edit.ID(), trackName).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Update releases to halted
	for _, r := range track.Releases {
		if r.Status == "inProgress" {
			r.Status = "halted"
		}
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would halt rollout on track '%s'", trackName)
		return output.Print(track)
	}

	updated, err := edit.Tracks().Update(client.GetPackageName(), edit.ID(), trackName, track).Context(ctx).Do()
	if err != nil {
		return err
	}

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Halted rollout on track '%s'", trackName)
	return output.Print(updated)
}

func runComplete(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	ctx := edit.Context()

	// Get current track
	track, err := edit.Tracks().Get(client.GetPackageName(), edit.ID(), trackName).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Update releases to completed
	for _, r := range track.Releases {
		if r.Status == "inProgress" || r.Status == "halted" {
			r.Status = "completed"
			r.UserFraction = 0 // Remove user fraction for full rollout
		}
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would complete rollout on track '%s'", trackName)
		return output.Print(track)
	}

	updated, err := edit.Tracks().Update(client.GetPackageName(), edit.ID(), trackName, track).Context(ctx).Do()
	if err != nil {
		return err
	}

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Completed rollout on track '%s' (100%%)", trackName)
	return output.Print(updated)
}
