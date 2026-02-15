package apps

import (
	"time"

	"github.com/spf13/cobra"

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
	Title       string `json:"title,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	// Note: The Google Play Developer API doesn't have a direct "list all apps" endpoint.
	// In a real implementation, you would need to iterate through known packages
	// or use the Play Console API (different from Publisher API).

	output.PrintInfo("To list apps, you need to know the package names.")
	output.PrintInfo("Use 'gpc apps get --package <package_name>' to verify access to a specific app.")

	return output.Print(map[string]string{
		"note": "Google Play Developer API requires specifying a package name. Use 'gpc apps get --package <name>' to verify access.",
	})
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
	defer edit.Delete() // Don't commit, just checking access

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
