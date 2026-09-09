package reports

import (
	"strings"

	"github.com/spf13/cobra"

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

// cliPlaceholder is replaced with the running binary's name in Access hints
const cliPlaceholder = "{cli}"

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
		Access:      cliPlaceholder + " vitals crashes",
	},
	{
		Type:        "anr",
		Description: "ANR (Application Not Responding) reports",
		Frequency:   "Real-time",
		Format:      "JSON via Vitals API",
		Access:      cliPlaceholder + " vitals anr",
	},
	{
		Type:        "reviews",
		Description: "User reviews and ratings",
		Frequency:   "Real-time",
		Format:      "JSON",
		Access:      cliPlaceholder + " reviews list",
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

// reportsFor returns the report list with Access hints resolved against the
// name the CLI was invoked as.
func reportsFor(cmd *cobra.Command) []ReportInfo {
	name := cmd.Root().Name()
	reports := make([]ReportInfo, 0, len(AvailableReports))
	for _, r := range AvailableReports {
		r.Access = strings.ReplaceAll(r.Access, cliPlaceholder, name)
		reports = append(reports, r)
	}
	return reports
}

func runList(cmd *cobra.Command, args []string) error {
	reports := reportsFor(cmd)

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
	return output.Print(reportsFor(cmd))
}
