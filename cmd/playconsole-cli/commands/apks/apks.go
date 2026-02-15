package apks

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

var APKsCmd = &cobra.Command{
	Use:   "apks",
	Short: "Manage APKs (legacy)",
	Long: `Upload and manage APK files.

Note: App Bundles (AAB) are the recommended format for publishing on Google Play.
APK support is maintained for legacy apps and specific use cases.`,
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an APK file",
	RunE:  runUpload,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List uploaded APKs",
	RunE:  runList,
}

var (
	filePath   string
	trackName  string
	autoCommit bool
)

func init() {
	uploadCmd.Flags().StringVar(&filePath, "file", "", "path to APK file")
	uploadCmd.Flags().StringVar(&trackName, "track", "", "track to assign")
	uploadCmd.Flags().BoolVar(&autoCommit, "commit", true, "automatically commit the edit")
	uploadCmd.MarkFlagRequired("file")

	APKsCmd.AddCommand(uploadCmd)
	APKsCmd.AddCommand(listCmd)
}

// APKInfo represents APK information
type APKInfo struct {
	VersionCode int64 `json:"version_code"`
	Binary      struct {
		SHA1   string `json:"sha1,omitempty"`
		SHA256 string `json:"sha256,omitempty"`
	} `json:"binary,omitempty"`
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
	if ext != ".apk" {
		output.PrintWarning("File does not have .apk extension: %s", ext)
	}

	output.PrintWarning("APK format is deprecated. Consider using App Bundles (AAB) instead.")

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would upload %s (%d bytes)", absPath, info.Size())
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 5*time.Minute)
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

	output.PrintInfo("Uploading APK: %s (%d bytes)", filepath.Base(absPath), info.Size())

	// Upload APK
	apk, err := edit.APKs().Upload(client.GetPackageName(), edit.ID()).Media(file).Context(edit.Context()).Do()
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	output.PrintSuccess("APK uploaded: version code %d", apk.VersionCode)

	// Assign to track if specified
	if trackName != "" {
		track := &androidpublisher.Track{
			Track: trackName,
			Releases: []*androidpublisher.TrackRelease{
				{
					VersionCodes: []int64{int64(apk.VersionCode)},
					Status:       "completed",
				},
			},
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
	}

	result := APKInfo{
		VersionCode: int64(apk.VersionCode),
	}
	if apk.Binary != nil {
		result.Binary.SHA1 = apk.Binary.Sha1
		result.Binary.SHA256 = apk.Binary.Sha256
	}

	return output.Print(result)
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

	apks, err := edit.APKs().List(client.GetPackageName(), edit.ID()).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	result := make([]APKInfo, 0, len(apks.Apks))
	for _, a := range apks.Apks {
		info := APKInfo{
			VersionCode: int64(a.VersionCode),
		}
		if a.Binary != nil {
			info.Binary.SHA1 = a.Binary.Sha1
			info.Binary.SHA256 = a.Binary.Sha256
		}
		result = append(result, info)
	}

	if len(result) == 0 {
		output.PrintInfo("No APKs found")
		return nil
	}

	return output.Print(result)
}
