package devices

import (
	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/cli"
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
	Short: "View device statistics overview",
	RunE:  runStats,
}

func init() {
	DevicesCmd.AddCommand(listCmd)
	DevicesCmd.AddCommand(statsCmd)
}

// DeviceInfo represents device information
type DeviceInfo struct {
	FormFactor  string `json:"form_factor"`
	Description string `json:"description"`
	Examples    string `json:"examples"`
}

// DeviceStats represents device usage statistics
type DeviceStats struct {
	PackageName      string   `json:"package_name"`
	TopManufacturers []string `json:"top_manufacturers"`
	TopFormFactors   []string `json:"top_form_factors"`
	Note             string   `json:"note"`
}

func runList(cmd *cobra.Command, args []string) error {
	devices := []DeviceInfo{
		{FormFactor: "phone", Description: "Smartphones", Examples: "Pixel, Samsung Galaxy, OnePlus"},
		{FormFactor: "tablet", Description: "Tablets", Examples: "Pixel Tablet, Samsung Tab, Lenovo Tab"},
		{FormFactor: "tv", Description: "Android TV", Examples: "Chromecast, Shield TV, Smart TVs"},
		{FormFactor: "wear", Description: "Wear OS watches", Examples: "Pixel Watch, Galaxy Watch"},
		{FormFactor: "auto", Description: "Android Auto", Examples: "Car head units"},
		{FormFactor: "chromebook", Description: "Chrome OS devices", Examples: "Chromebooks with Play Store"},
	}

	return output.Print(devices)
}

func runStats(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	stats := DeviceStats{
		PackageName:      cli.GetPackageName(),
		TopManufacturers: []string{"Samsung", "Xiaomi", "OPPO", "vivo", "Google", "OnePlus", "Huawei", "Motorola"},
		TopFormFactors:   []string{"phone", "tablet"},
		Note:             "Detailed device stats available in Play Console web interface",
	}

	return output.Print(stats)
}
