package bundles

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var BundlesCmd = &cobra.Command{
	Use:   "bundles",
	Short: "Manage Android App Bundles",
	Long: `Upload and manage Android App Bundles (AAB files).

App Bundles are the recommended format for publishing on Google Play.
They allow for smaller downloads and dynamic feature delivery.`,
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an Android App Bundle",
	Long: `Upload an Android App Bundle (AAB) to Google Play.

After uploading, the bundle can be assigned to a track using 'gpc tracks update'.`,
	RunE: runUpload,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List uploaded bundles",
	RunE:  runList,
}

var (
	filePath    string
	trackName   string
	autoCommit  bool
	releaseNotes string
	releaseNotesLang string
	rolloutPct  float64
)

func init() {
	// Upload flags
	uploadCmd.Flags().StringVar(&filePath, "file", "", "path to AAB file")
	uploadCmd.Flags().StringVar(&trackName, "track", "", "track to assign (internal, alpha, beta, production)")
	uploadCmd.Flags().BoolVar(&autoCommit, "commit", true, "automatically commit the edit")
	uploadCmd.Flags().StringVar(&releaseNotes, "release-notes", "", "release notes text")
	uploadCmd.Flags().StringVar(&releaseNotesLang, "release-notes-lang", "en-US", "release notes language")
	uploadCmd.Flags().Float64Var(&rolloutPct, "rollout", 100, "rollout percentage (only for production)")
	uploadCmd.MarkFlagRequired("file")

	BundlesCmd.AddCommand(uploadCmd)
	BundlesCmd.AddCommand(listCmd)
}

// BundleInfo represents bundle information
type BundleInfo struct {
	VersionCode int64  `json:"version_code"`
	SHA1        string `json:"sha1,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
}

func runUpload(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	// Validate file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("file not found: %s", absPath)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", absPath)
	}

	ext := filepath.Ext(absPath)
	if ext != ".aab" {
		output.PrintWarning("File does not have .aab extension: %s", ext)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would upload %s (%d bytes)", absPath, info.Size())
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 5*time.Minute) // Longer timeout for uploads
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	// Open file
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	output.PrintInfo("Uploading bundle: %s (%d bytes)", filepath.Base(absPath), info.Size())

	// Upload bundle
	bundle, err := edit.Bundles().Upload(client.GetPackageName(), edit.ID()).Media(file).Context(edit.Context()).Do()
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	output.PrintSuccess("Bundle uploaded: version code %d", bundle.VersionCode)

	// Assign to track if specified
	if trackName != "" {
		release := &androidpublisher.TrackRelease{
			VersionCodes: []int64{bundle.VersionCode},
			Status:       "completed",
		}

		// Handle staged rollout
		if rolloutPct < 100 {
			release.Status = "inProgress"
			release.UserFraction = rolloutPct / 100
		}

		// Add release notes
		if releaseNotes != "" {
			release.ReleaseNotes = []*androidpublisher.LocalizedText{
				{
					Language: releaseNotesLang,
					Text:     releaseNotes,
				},
			}
		}

		track := &androidpublisher.Track{
			Track:    trackName,
			Releases: []*androidpublisher.TrackRelease{release},
		}

		_, err := edit.Tracks().Update(client.GetPackageName(), edit.ID(), trackName, track).Context(edit.Context()).Do()
		if err != nil {
			return fmt.Errorf("failed to assign to track '%s': %w", trackName, err)
		}

		output.PrintSuccess("Assigned to track: %s", trackName)
	}

	// Commit if requested
	if autoCommit {
		if err := edit.Commit(); err != nil {
			return err
		}
		output.PrintSuccess("Edit committed")
	} else {
		output.PrintInfo("Edit ID: %s (not committed, use 'gpc edits commit' to commit)", edit.ID())
	}

	return output.Print(BundleInfo{
		VersionCode: bundle.VersionCode,
		SHA1:        bundle.Sha1,
		SHA256:      bundle.Sha256,
	})
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

	bundles, err := edit.Bundles().List(client.GetPackageName(), edit.ID()).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	result := make([]BundleInfo, 0, len(bundles.Bundles))
	for _, b := range bundles.Bundles {
		result = append(result, BundleInfo{
			VersionCode: b.VersionCode,
			SHA1:        b.Sha1,
			SHA256:      b.Sha256,
		})
	}

	if len(result) == 0 {
		output.PrintInfo("No bundles found")
		return nil
	}

	return output.Print(result)
}
