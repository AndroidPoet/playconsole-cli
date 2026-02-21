package reports

import (
	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var ReportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "View available reports",
	Long: `View information about available reports from Google Play Console.

Note: Full report downloads require the Play Console web interface or
Cloud Storage export. This command shows report types and availability.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available report types",
	RunE:  runList,
}

var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "Show all report types with descriptions",
	RunE:  runTypes,
}

var (
	reportType string
)

func init() {
	listCmd.Flags().StringVar(&reportType, "type", "", "filter by report type")

	ReportsCmd.AddCommand(listCmd)
	ReportsCmd.AddCommand(typesCmd)
}

// ReportInfo represents report information
type ReportInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Frequency   string `json:"frequency"`
	Format      string `json:"format"`
	Access      string `json:"access"`
}

// AvailableReports lists all available report types
var AvailableReports = []ReportInfo{
	{
		Type:        "installs",
		Description: "Daily install and uninstall statistics",
		Frequency:   "Daily",
		Format:      "CSV",
		Access:      "Cloud Storage export",
	},
	{
		Type:        "crashes",
		Description: "Crash reports with stack traces",
		Frequency:   "Real-time",
		Format:      "JSON via Vitals API",
		Access:      "gpc vitals crashes",
	},
	{
		Type:        "anr",
		Description: "ANR (Application Not Responding) reports",
		Frequency:   "Real-time",
		Format:      "JSON via Vitals API",
		Access:      "gpc vitals anr",
	},
	{
		Type:        "reviews",
		Description: "User reviews and ratings",
		Frequency:   "Real-time",
		Format:      "JSON",
		Access:      "gpc reviews list",
	},
	{
		Type:        "ratings",
		Description: "Rating distribution over time",
		Frequency:   "Daily",
		Format:      "CSV",
		Access:      "Cloud Storage export",
	},
	{
		Type:        "financial",
		Description: "Earnings and financial reports",
		Frequency:   "Monthly",
		Format:      "CSV",
		Access:      "Play Console (requires merchant)",
	},
	{
		Type:        "subscriptions",
		Description: "Subscription metrics and churn",
		Frequency:   "Daily",
		Format:      "CSV",
		Access:      "Cloud Storage export",
	},
	{
		Type:        "statistics",
		Description: "Aggregate app statistics",
		Frequency:   "Daily",
		Format:      "CSV",
		Access:      "Cloud Storage export",
	},
}

func runList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	reports := AvailableReports

	// Filter by type if specified
	if reportType != "" {
		filtered := []ReportInfo{}
		for _, r := range reports {
			if r.Type == reportType {
				filtered = append(filtered, r)
			}
		}
		reports = filtered
	}

	return output.Print(reports)
}

func runTypes(cmd *cobra.Command, args []string) error {
	return output.Print(AvailableReports)
}
