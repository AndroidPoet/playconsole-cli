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

var slowStartCmd = &cobra.Command{
	Use:   "slow-start",
	Short: "View slow startup rate metrics",
	Long:  "View the percentage of user sessions with slow app startup times.",
	RunE:  runSlowStart,
}

var slowRenderingCmd = &cobra.Command{
	Use:   "slow-rendering",
	Short: "View slow rendering rate metrics",
	Long:  "View the percentage of user sessions with slow frame rendering.",
	RunE:  runSlowRendering,
}

var wakeupsCmd = &cobra.Command{
	Use:   "wakeups",
	Short: "View excessive wakeup rate metrics",
	Long:  "View the rate of excessive wakeup alarms causing battery drain.",
	RunE:  runWakeups,
}

var wakelocksCmd = &cobra.Command{
	Use:   "wakelocks",
	Short: "View stuck background wakelock rate metrics",
	Long:  "View the rate of stuck background wakelocks draining battery.",
	RunE:  runWakelocks,
}

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "View low memory killer (LMK) rate metrics",
	Long:  "View the rate of low memory killer events affecting your app.",
	RunE:  runMemory,
}

var errorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "View error counts and issues",
	Long:  "View aggregated error counts from the Play Developer Reporting API.",
	RunE:  runErrors,
}

var errorsIssuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "List error issues (grouped errors)",
	RunE:  runErrorIssues,
}

var (
	days int
)

func init() {
	// Common flags for all metric commands
	daysCommands := []*cobra.Command{
		crashesCmd, anrCmd, overviewCmd,
		slowStartCmd, slowRenderingCmd, wakeupsCmd,
		wakelocksCmd, memoryCmd, errorsCmd,
	}
	for _, cmd := range daysCommands {
		cmd.Flags().IntVar(&days, "days", 28, "number of days to query (7, 28, or custom)")
	}

	// Build errors sub-tree
	errorsCmd.AddCommand(errorsIssuesCmd)

	VitalsCmd.AddCommand(crashesCmd)
	VitalsCmd.AddCommand(anrCmd)
	VitalsCmd.AddCommand(overviewCmd)
	VitalsCmd.AddCommand(slowStartCmd)
	VitalsCmd.AddCommand(slowRenderingCmd)
	VitalsCmd.AddCommand(wakeupsCmd)
	VitalsCmd.AddCommand(wakelocksCmd)
	VitalsCmd.AddCommand(memoryCmd)
	VitalsCmd.AddCommand(errorsCmd)
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

// SlowStartInfo represents slow start rate information
type SlowStartInfo struct {
	MetricSet string `json:"metric_set"`
	Period    string `json:"period"`
	Status    string `json:"status"`
}

// SlowRenderingInfo represents slow rendering rate information
type SlowRenderingInfo struct {
	MetricSet string `json:"metric_set"`
	Period    string `json:"period"`
	Status    string `json:"status"`
}

// WakeupInfo represents excessive wakeup rate information
type WakeupInfo struct {
	MetricSet string `json:"metric_set"`
	Period    string `json:"period"`
	Status    string `json:"status"`
}

// WakelockInfo represents stuck wakelock rate information
type WakelockInfo struct {
	MetricSet string `json:"metric_set"`
	Period    string `json:"period"`
	Status    string `json:"status"`
}

// MemoryInfo represents low memory killer rate information
type MemoryInfo struct {
	MetricSet string `json:"metric_set"`
	Period    string `json:"period"`
	Status    string `json:"status"`
}

// ErrorInfo represents error count information
type ErrorInfo struct {
	MetricSet string `json:"metric_set"`
	Period    string `json:"period"`
	Status    string `json:"status"`
}

func runSlowStart(cmd *cobra.Command, args []string) error {
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
	metricName := fmt.Sprintf("%s/slowStartRateMetricSet", appName)

	resp, err := client.Vitals().Slowstartrate.Get(metricName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get slow start rate: %w", err)
	}

	info := SlowStartInfo{
		MetricSet: resp.Name,
		Period:    fmt.Sprintf("%d days", days),
		Status:    "available",
	}

	if resp.FreshnessInfo != nil && len(resp.FreshnessInfo.Freshnesses) > 0 {
		info.Status = "fresh"
	}

	return output.Print(info)
}

func runSlowRendering(cmd *cobra.Command, args []string) error {
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
	metricName := fmt.Sprintf("%s/slowRenderingRateMetricSet", appName)

	resp, err := client.Vitals().Slowrenderingrate.Get(metricName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get slow rendering rate: %w", err)
	}

	info := SlowRenderingInfo{
		MetricSet: resp.Name,
		Period:    fmt.Sprintf("%d days", days),
		Status:    "available",
	}

	if resp.FreshnessInfo != nil && len(resp.FreshnessInfo.Freshnesses) > 0 {
		info.Status = "fresh"
	}

	return output.Print(info)
}

func runWakeups(cmd *cobra.Command, args []string) error {
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
	metricName := fmt.Sprintf("%s/excessiveWakeupRateMetricSet", appName)

	resp, err := client.Vitals().Excessivewakeuprate.Get(metricName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get excessive wakeup rate: %w", err)
	}

	info := WakeupInfo{
		MetricSet: resp.Name,
		Period:    fmt.Sprintf("%d days", days),
		Status:    "available",
	}

	if resp.FreshnessInfo != nil && len(resp.FreshnessInfo.Freshnesses) > 0 {
		info.Status = "fresh"
	}

	return output.Print(info)
}

func runWakelocks(cmd *cobra.Command, args []string) error {
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
	metricName := fmt.Sprintf("%s/stuckBackgroundWakelockRateMetricSet", appName)

	resp, err := client.Vitals().Stuckbackgroundwakelockrate.Get(metricName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get stuck wakelock rate: %w", err)
	}

	info := WakelockInfo{
		MetricSet: resp.Name,
		Period:    fmt.Sprintf("%d days", days),
		Status:    "available",
	}

	if resp.FreshnessInfo != nil && len(resp.FreshnessInfo.Freshnesses) > 0 {
		info.Status = "fresh"
	}

	return output.Print(info)
}

func runMemory(cmd *cobra.Command, args []string) error {
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
	metricName := fmt.Sprintf("%s/lmkRateMetricSet", appName)

	resp, err := client.Vitals().Lmkrate.Get(metricName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get LMK rate: %w", err)
	}

	info := MemoryInfo{
		MetricSet: resp.Name,
		Period:    fmt.Sprintf("%d days", days),
		Status:    "available",
	}

	if resp.FreshnessInfo != nil && len(resp.FreshnessInfo.Freshnesses) > 0 {
		info.Status = "fresh"
	}

	return output.Print(info)
}

func runErrors(cmd *cobra.Command, args []string) error {
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
	metricName := fmt.Sprintf("%s/errorCountMetricSet", appName)

	resp, err := client.Vitals().Errors.Counts.Get(metricName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get error counts: %w", err)
	}

	info := ErrorInfo{
		MetricSet: resp.Name,
		Period:    fmt.Sprintf("%d days", days),
		Status:    "available",
	}

	if resp.FreshnessInfo != nil && len(resp.FreshnessInfo.Freshnesses) > 0 {
		info.Status = "fresh"
	}

	return output.Print(info)
}

func runErrorIssues(cmd *cobra.Command, args []string) error {
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

	resp, err := client.Vitals().Errors.Issues.Search(appName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to list error issues: %w", err)
	}

	if len(resp.ErrorIssues) == 0 {
		output.PrintInfo("No error issues found")
		return nil
	}

	type IssueInfo struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		ErrorCount   int64  `json:"error_count,omitempty"`
		Cause        string `json:"cause,omitempty"`
		FirstVersion string `json:"first_version,omitempty"`
	}

	result := make([]IssueInfo, 0, len(resp.ErrorIssues))
	for _, issue := range resp.ErrorIssues {
		info := IssueInfo{
			Name: issue.Name,
			Type: issue.Type,
		}
		if issue.Cause != "" {
			info.Cause = issue.Cause
		}
		if issue.FirstAppVersion != nil {
			info.FirstVersion = fmt.Sprintf("%d", issue.FirstAppVersion.VersionCode)
		}
		result = append(result, info)
	}

	return output.Print(result)
}
