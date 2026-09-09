package testing

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var TestingCmd = &cobra.Command{
	Use:   "testing",
	Short: "Manage testing tracks and testers",
	Long: `Manage internal testing, closed testing, and open testing.

This allows you to upload builds for internal sharing, manage tester
groups, and control access to test builds.`,
}

var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "Internal testing management",
}

var internalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List internal test builds",
	RunE:  runInternalList,
}

var internalSharingCmd = &cobra.Command{
	Use:   "internal-sharing",
	Short: "Internal app sharing",
}

var internalSharingUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a build for internal sharing",
	RunE:  runInternalSharingUpload,
}

var testersCmd = &cobra.Command{
	Use:   "testers",
	Short: "Manage testers",
	Long: `Manage the Google Groups assigned as testers on a track.

The Play Developer API only accepts Google Group email addresses here;
individual tester emails are not supported by the API and must be managed
through the Play Console UI or by adding the user to a Google Group.`,
}

var testersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tester Google Groups for a track",
	RunE:  runTestersList,
}

var testersAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add tester Google Groups to a track",
	Long: `Add Google Group email addresses as testers on a track.

Individual tester emails are not supported by the API; pass Google Group
addresses (for example testers@googlegroups.com).`,
	RunE: runTestersAdd,
}

var testersRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove tester Google Groups from a track",
	Long: `Remove Google Group email addresses from a track's testers.

Individual tester emails are not supported by the API; pass the Google Group
addresses currently assigned to the track.`,
	RunE: runTestersRemove,
}

var testerGroupsCmd = &cobra.Command{
	Use:   "tester-groups",
	Short: "Manage tester groups",
}

var testerGroupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tester groups",
	Long: `The Play Developer API has no endpoint for listing tester groups.

Groups are Google Groups managed outside the API; use
'gpc testing testers list --track <track>' to see the groups assigned to a track.`,
	RunE: runTesterGroupsList,
}

var (
	trackName  string
	emails     string
	emailsFile string
	filePath   string
)

func init() {
	// Testers list flags
	testersListCmd.Flags().StringVar(&trackName, "track", "", "track name (alpha, beta, etc.)")
	cli.MustMarkFlagRequired(testersListCmd, "track")

	// Testers add flags
	testersAddCmd.Flags().StringVar(&trackName, "track", "", "track name")
	testersAddCmd.Flags().StringVar(&emails, "emails", "", "comma-separated Google Group email addresses (individual tester emails are not supported by the API)")
	testersAddCmd.Flags().StringVar(&emailsFile, "emails-file", "", "file containing Google Group email addresses, one per line (individual tester emails are not supported by the API)")
	cli.MustMarkFlagRequired(testersAddCmd, "track")

	// Testers remove flags
	testersRemoveCmd.Flags().StringVar(&trackName, "track", "", "track name")
	testersRemoveCmd.Flags().StringVar(&emails, "emails", "", "comma-separated Google Group email addresses (individual tester emails are not supported by the API)")
	testersRemoveCmd.Flags().Bool("confirm", false, "confirm removal")
	cli.MustMarkFlagRequired(testersRemoveCmd, "track")
	cli.MustMarkFlagRequired(testersRemoveCmd, "emails")

	// Internal sharing upload flags
	internalSharingUploadCmd.Flags().StringVar(&filePath, "file", "", "path to APK/AAB file")
	cli.MustMarkFlagRequired(internalSharingUploadCmd, "file")

	// Build command tree
	internalCmd.AddCommand(internalListCmd)
	internalSharingCmd.AddCommand(internalSharingUploadCmd)

	testersCmd.AddCommand(testersListCmd)
	testersCmd.AddCommand(testersAddCmd)
	testersCmd.AddCommand(testersRemoveCmd)

	testerGroupsCmd.AddCommand(testerGroupsListCmd)

	TestingCmd.AddCommand(internalCmd)
	TestingCmd.AddCommand(internalSharingCmd)
	TestingCmd.AddCommand(testersCmd)
	TestingCmd.AddCommand(testerGroupsCmd)
}

func runInternalList(cmd *cobra.Command, args []string) error {
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

	// Get the internal track
	track, err := edit.Tracks().Get(client.GetPackageName(), edit.ID(), "internal").Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	type ReleaseInfo struct {
		VersionCodes []int64 `json:"version_codes"`
		Status       string  `json:"status"`
		Name         string  `json:"name,omitempty"`
	}

	result := make([]ReleaseInfo, 0, len(track.Releases))
	for _, r := range track.Releases {
		result = append(result, ReleaseInfo{
			VersionCodes: r.VersionCodes,
			Status:       r.Status,
			Name:         r.Name,
		})
	}

	if len(result) == 0 {
		output.PrintInfo("No internal test releases found")
	}

	return output.Print(result)
}

func runInternalSharingUpload(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("file not found: %s", absPath)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would upload %s for internal sharing", filepath.Base(absPath))
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 5*time.Minute)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	output.PrintInfo("Uploading for internal sharing: %s (%d bytes)", filepath.Base(absPath), info.Size())

	// Determine if AAB or APK
	ext := strings.ToLower(filepath.Ext(absPath))
	var downloadURL string

	if ext == ".aab" {
		artifact, err := client.GetService().Internalappsharingartifacts.Uploadbundle(client.GetPackageName()).Media(file, googleapi.ContentType("application/octet-stream")).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		downloadURL = artifact.DownloadUrl
	} else {
		artifact, err := client.GetService().Internalappsharingartifacts.Uploadapk(client.GetPackageName()).Media(file, googleapi.ContentType("application/octet-stream")).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		downloadURL = artifact.DownloadUrl
	}

	output.PrintSuccess("Upload complete")
	return output.Print(map[string]interface{}{
		"download_url": downloadURL,
		"file":         filepath.Base(absPath),
	})
}

func runTestersList(cmd *cobra.Command, args []string) error {
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

	testers, err := edit.Testers().Get(client.GetPackageName(), edit.ID(), trackName).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	groups := testers.GoogleGroups
	if len(groups) == 0 {
		output.PrintInfo("No testers configured for track '%s'", trackName)
		groups = []string{}
	}

	return output.Print(map[string]interface{}{
		"track":         trackName,
		"google_groups": groups,
	})
}

func runTestersAdd(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	// Collect emails
	emailList := splitEmails(emails)

	if emailsFile != "" {
		data, err := os.ReadFile(emailsFile)
		if err != nil {
			return fmt.Errorf("failed to read emails file: %w", err)
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				emailList = append(emailList, line)
			}
		}
	}

	if len(emailList) == 0 {
		return fmt.Errorf("no Google Group emails provided. Use --emails or --emails-file")
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would add %d tester groups to track '%s'", len(emailList), trackName)
		return output.Print(emailList)
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

	// Get existing testers
	existing, err := edit.Testers().Get(client.GetPackageName(), edit.ID(), trackName).Context(ctx).Do()
	if err != nil {
		// Only a 404 means no testers exist yet; any other error must not
		// lead to replacing the list with a partial one.
		var gerr *googleapi.Error
		if !errors.As(err, &gerr) || gerr.Code != http.StatusNotFound {
			return fmt.Errorf("failed to get existing testers: %w", err)
		}
		existing = &androidpublisher.Testers{}
	}

	// Add new emails (avoiding duplicates, including within the input)
	existingMap := make(map[string]bool)
	for _, g := range existing.GoogleGroups {
		existingMap[g] = true
	}

	added := 0
	for _, email := range emailList {
		if !existingMap[email] {
			existing.GoogleGroups = append(existing.GoogleGroups, email)
			existingMap[email] = true
			added++
		}
	}

	// Update testers
	_, err = edit.Testers().Update(client.GetPackageName(), edit.ID(), trackName, existing).Context(ctx).Do()
	if err != nil {
		return err
	}

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Added %d tester groups to track '%s'", added, trackName)
	return nil
}

func runTestersRemove(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	emailList := splitEmails(emails)

	if len(emailList) == 0 {
		return fmt.Errorf("no Google Group emails provided. Use --emails")
	}

	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would remove %d tester groups from track '%s'", len(emailList), trackName)
		return output.Print(emailList)
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

	existing, err := edit.Testers().Get(client.GetPackageName(), edit.ID(), trackName).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Remove specified emails
	removeMap := make(map[string]bool)
	for _, e := range emailList {
		removeMap[e] = true
	}

	newGroups := make([]string, 0)
	removed := 0
	for _, g := range existing.GoogleGroups {
		if removeMap[g] {
			removed++
		} else {
			newGroups = append(newGroups, g)
		}
	}

	if removed == 0 {
		return fmt.Errorf("none of the given Google Groups are assigned to track '%s'", trackName)
	}

	existing.GoogleGroups = newGroups

	_, err = edit.Testers().Update(client.GetPackageName(), edit.ID(), trackName, existing).Context(ctx).Do()
	if err != nil {
		return err
	}

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Removed %d tester groups from track '%s'", removed, trackName)
	return nil
}

func runTesterGroupsList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	// Tester groups are Google Groups managed outside the API; the API only
	// exposes the groups assigned to a track.
	output.PrintInfo("The Play Developer API has no endpoint for listing tester groups.")
	output.PrintInfo("Use 'gpc testing testers list --track <track>' to see assigned groups.")
	return output.Print(map[string]interface{}{
		"supported": false,
		"message":   "The Play Developer API has no endpoint for listing tester groups; use 'testing testers list --track <track>' to see the Google Groups assigned to a track",
		"groups":    []string{},
	})
}

// splitEmails splits a comma-separated list, trimming whitespace and
// dropping empty entries (for example from a trailing comma).
func splitEmails(s string) []string {
	var result []string
	for _, e := range strings.Split(s, ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			result = append(result, e)
		}
	}
	return result
}
