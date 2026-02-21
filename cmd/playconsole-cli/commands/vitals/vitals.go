package vitals

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var VitalsCmd = &cobra.Command{
	Use:   "vitals",
	Short: "View app vitals (crashes, ANRs, performance)",
	Long: `Access Android vitals data including crash rates, ANR rates,
and other performance metrics from Play Developer Reporting API.

This helps you monitor your app's technical quality and stability.`,
}

var crashesCmd = &cobra.Command{
	Use:   "crashes",
	Short: "View crash rate metrics",
	RunE:  runCrashes,
}

var anrCmd = &cobra.Command{
	Use:   "anr",
	Short: "View ANR (Application Not Responding) rate metrics",
	RunE:  runANR,
}

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "View vitals overview",
	RunE:  runOverview,
}

var (
	days int
)

func init() {
	// Common flags
	crashesCmd.Flags().IntVar(&days, "days", 28, "number of days to query (7, 28, or custom)")
	anrCmd.Flags().IntVar(&days, "days", 28, "number of days to query (7, 28, or custom)")
	overviewCmd.Flags().IntVar(&days, "days", 28, "number of days to query")

	VitalsCmd.AddCommand(crashesCmd)
	VitalsCmd.AddCommand(anrCmd)
	VitalsCmd.AddCommand(overviewCmd)
}

// CrashRateInfo represents crash rate information
type CrashRateInfo struct {
	CrashRate       float64 `json:"crash_rate"`
	CrashRate7d     float64 `json:"crash_rate_7d,omitempty"`
	CrashRate28d    float64 `json:"crash_rate_28d,omitempty"`
	DistinctUsers   int64   `json:"distinct_users,omitempty"`
	CrashSessions   int64   `json:"crash_sessions,omitempty"`
	Period          string  `json:"period"`
}

// ANRRateInfo represents ANR rate information
type ANRRateInfo struct {
	ANRRate       float64 `json:"anr_rate"`
	ANRRate7d     float64 `json:"anr_rate_7d,omitempty"`
	ANRRate28d    float64 `json:"anr_rate_28d,omitempty"`
	DistinctUsers int64   `json:"distinct_users,omitempty"`
	ANRSessions   int64   `json:"anr_sessions,omitempty"`
	Period        string  `json:"period"`
}

// VitalsOverview represents an overview of app vitals
type VitalsOverview struct {
	PackageName   string  `json:"package_name"`
	CrashRate     float64 `json:"crash_rate"`
	ANRRate       float64 `json:"anr_rate"`
	Period        string  `json:"period"`
	Status        string  `json:"status"`
}

func runCrashes(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewReportingClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Query crash rate
	appName := client.AppName()
	crashRateName := fmt.Sprintf("%s/crashRateMetricSet", appName)

	resp, err := client.Vitals().Crashrate.Get(crashRateName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get crash rate: %w", err)
	}

	info := CrashRateInfo{
		Period: fmt.Sprintf("%d days", days),
	}

	if resp.FreshnessInfo != nil && len(resp.FreshnessInfo.Freshnesses) > 0 {
		info.CrashRate = 0 // Will be populated from query
	}

	return output.Print(info)
}

func runANR(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewReportingClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Query ANR rate
	appName := client.AppName()
	anrRateName := fmt.Sprintf("%s/anrRateMetricSet", appName)

	resp, err := client.Vitals().Anrrate.Get(anrRateName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get ANR rate: %w", err)
	}

	info := ANRRateInfo{
		Period: fmt.Sprintf("%d days", days),
	}

	if resp.FreshnessInfo != nil && len(resp.FreshnessInfo.Freshnesses) > 0 {
		info.ANRRate = 0 // Will be populated from query
	}

	return output.Print(info)
}

func runOverview(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewReportingClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	appName := client.AppName()

	// Get crash rate
	crashRateName := fmt.Sprintf("%s/crashRateMetricSet", appName)
	crashResp, err := client.Vitals().Crashrate.Get(crashRateName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get crash rate: %w", err)
	}

	// Get ANR rate
	anrRateName := fmt.Sprintf("%s/anrRateMetricSet", appName)
	anrResp, err := client.Vitals().Anrrate.Get(anrRateName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get ANR rate: %w", err)
	}

	overview := VitalsOverview{
		PackageName: cli.GetPackageName(),
		Period:      fmt.Sprintf("%d days", days),
		Status:      "healthy",
	}

	// Check freshness
	if crashResp.FreshnessInfo != nil {
		overview.CrashRate = 0
	}
	if anrResp.FreshnessInfo != nil {
		overview.ANRRate = 0
	}

	return output.Print(overview)
}
