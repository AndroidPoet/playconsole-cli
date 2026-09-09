package apps

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var AppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage applications",
	Long:  `List and manage applications in your Google Play Console account.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	Long:  `List all applications you have access to in Google Play Console.`,
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get application details",
	RunE:  runGet,
}

func init() {
	AppsCmd.AddCommand(listCmd)
	AppsCmd.AddCommand(getCmd)
}

// AppInfo represents basic app information
type AppInfo struct {
	PackageName string `json:"package_name"`
	DisplayName string `json:"display_name,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	var apps []AppInfo

	err := api.CheckAndEnableAPI(func() error {
		result, err := fetchApps()
		if err != nil {
			return err
		}
		apps = result
		return nil
	})

	if err != nil {
		return err
	}

	if len(apps) == 0 {
		output.PrintInfo("No apps found. Make sure your service account has access to apps in Play Console.")
		return output.Print([]AppInfo{})
	}

	return output.Print(apps)
}

func fetchApps() ([]AppInfo, error) {
	// The search is account-wide, so no package name is needed. Going through
	// the shared reporting client keeps --debug and --timeout handling uniform.
	client, err := api.NewReportingClient("", 60*time.Second)
	if err != nil {
		return nil, err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Search for all accessible apps, following every page
	apps := make([]AppInfo, 0)
	err = client.Apps().Search().Context(ctx).Pages(ctx,
		func(response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1SearchAccessibleAppsResponse) error {
			for _, app := range response.Apps {
				// Resource name format: apps/{package_name}
				packageName := strings.TrimPrefix(app.Name, "apps/")
				apps = append(apps, AppInfo{
					PackageName: packageName,
					DisplayName: app.DisplayName,
				})
			}
			return nil
		})
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func runGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	// Create a temporary edit to verify access and get app info
	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	// Get app details
	ctx := edit.Context()
	details, err := edit.Details().Get(client.GetPackageName(), edit.ID()).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Get available tracks for additional info
	tracks, err := edit.Tracks().List(client.GetPackageName(), edit.ID()).Context(ctx).Do()
	if err != nil {
		// Non-fatal, just means we can't get track info
		tracks = nil
	}

	result := map[string]interface{}{
		"package_name":     client.GetPackageName(),
		"default_language": details.DefaultLanguage,
		"contact_email":    details.ContactEmail,
		"contact_phone":    details.ContactPhone,
		"contact_website":  details.ContactWebsite,
	}

	if tracks != nil {
		trackNames := make([]string, 0, len(tracks.Tracks))
		for _, t := range tracks.Tracks {
			trackNames = append(trackNames, t.Track)
		}
		result["tracks"] = trackNames
	}

	return output.Print(result)
}
