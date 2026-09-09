package devices

import (
	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var DevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "View device catalog and compatibility",
	Long: `View the device catalog and check which devices are compatible
with your app based on its requirements.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List common Android device form factors",
	RunE:  runList,
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View general Android device landscape (static reference data)",
	RunE:  runStats,
}

func init() {
	DevicesCmd.AddCommand(listCmd)
	DevicesCmd.AddCommand(statsCmd)
}

// sourceStatic marks output that is bundled reference data, not fetched
// from the Play Console for the current package.
const sourceStatic = "static"

// DeviceInfo represents device information
type DeviceInfo struct {
	FormFactor  string `json:"form_factor"`
	Description string `json:"description"`
	Examples    string `json:"examples"`
	Source      string `json:"source"`
}

// DeviceStats represents general device landscape reference data
type DeviceStats struct {
	Source           string   `json:"source"`
	TopManufacturers []string `json:"top_manufacturers"`
	TopFormFactors   []string `json:"top_form_factors"`
	Note             string   `json:"note"`
}

func warnStatic() {
	output.PrintWarning("the Reporting API has no device distribution metric; this is static reference data, not your app's real device distribution. See Play Console > Reach and devices.")
}

func runList(cmd *cobra.Command, args []string) error {
	warnStatic()

	devices := []DeviceInfo{
		{FormFactor: "phone", Description: "Smartphones", Examples: "Pixel, Samsung Galaxy, OnePlus", Source: sourceStatic},
		{FormFactor: "tablet", Description: "Tablets", Examples: "Pixel Tablet, Samsung Tab, Lenovo Tab", Source: sourceStatic},
		{FormFactor: "tv", Description: "Android TV", Examples: "Chromecast, Shield TV, Smart TVs", Source: sourceStatic},
		{FormFactor: "wear", Description: "Wear OS watches", Examples: "Pixel Watch, Galaxy Watch", Source: sourceStatic},
		{FormFactor: "auto", Description: "Android Auto", Examples: "Car head units", Source: sourceStatic},
		{FormFactor: "chromebook", Description: "Chrome OS devices", Examples: "Chromebooks with Play Store", Source: sourceStatic},
	}

	return output.Print(devices)
}

func runStats(cmd *cobra.Command, args []string) error {
	warnStatic()

	stats := DeviceStats{
		Source:           sourceStatic,
		TopManufacturers: []string{"Samsung", "Xiaomi", "OPPO", "vivo", "Google", "OnePlus", "Huawei", "Motorola"},
		TopFormFactors:   []string{"phone", "tablet"},
		Note:             "Static reference data. Per-app device stats are only available in the Play Console web interface",
	}

	return output.Print(stats)
}
