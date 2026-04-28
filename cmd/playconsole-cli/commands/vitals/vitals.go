package vitals

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

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
	daysCommands := []*cobra.Command{
		crashesCmd, anrCmd, overviewCmd,
		slowStartCmd, slowRenderingCmd, wakeupsCmd,
		wakelocksCmd, memoryCmd, errorsCmd,
	}
	for _, cmd := range daysCommands {
		cmd.Flags().IntVar(&days, "days", 28, "number of days to query (7, 28, or custom)")
	}

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

type CrashRateInfo struct {
	CrashRate     float64 `json:"crash_rate"`
	CrashRate7d   float64 `json:"crash_rate_7d,omitempty"`
	CrashRate28d  float64 `json:"crash_rate_28d,omitempty"`
	DistinctUsers float64 `json:"distinct_users,omitempty"`
	Period        string  `json:"period"`
}

type ANRRateInfo struct {
	ANRRate              float64 `json:"anr_rate"`
	ANRRate7d            float64 `json:"anr_rate_7d,omitempty"`
	ANRRate28d           float64 `json:"anr_rate_28d,omitempty"`
	UserPerceivedANRRate float64 `json:"user_perceived_anr_rate,omitempty"`
	DistinctUsers        float64 `json:"distinct_users,omitempty"`
	Period               string  `json:"period"`
}

type VitalsOverview struct {
	PackageName   string  `json:"package_name"`
	CrashRate     float64 `json:"crash_rate"`
	ANRRate       float64 `json:"anr_rate"`
	SlowStartRate float64 `json:"slow_start_rate,omitempty"`
	Period        string  `json:"period"`
}

type SlowStartInfo struct {
	SlowStartRate    float64 `json:"slow_start_rate"`
	SlowStartRate7d  float64 `json:"slow_start_rate_7d,omitempty"`
	SlowStartRate28d float64 `json:"slow_start_rate_28d,omitempty"`
	DistinctUsers    float64 `json:"distinct_users,omitempty"`
	Period           string  `json:"period"`
}

type SlowRenderingInfo struct {
	SlowRenderingRate20Fps    float64 `json:"slow_rendering_rate_20_fps,omitempty"`
	SlowRenderingRate20Fps7d  float64 `json:"slow_rendering_rate_20_fps_7d,omitempty"`
	SlowRenderingRate20Fps28d float64 `json:"slow_rendering_rate_20_fps_28d,omitempty"`
	SlowRenderingRate30Fps    float64 `json:"slow_rendering_rate_30_fps,omitempty"`
	SlowRenderingRate30Fps7d  float64 `json:"slow_rendering_rate_30_fps_7d,omitempty"`
	SlowRenderingRate30Fps28d float64 `json:"slow_rendering_rate_30_fps_28d,omitempty"`
	DistinctUsers             float64 `json:"distinct_users,omitempty"`
	Period                    string  `json:"period"`
}

type WakeupInfo struct {
	ExcessiveWakeupRate    float64 `json:"excessive_wakeup_rate"`
	ExcessiveWakeupRate7d  float64 `json:"excessive_wakeup_rate_7d,omitempty"`
	ExcessiveWakeupRate28d float64 `json:"excessive_wakeup_rate_28d,omitempty"`
	DistinctUsers          float64 `json:"distinct_users,omitempty"`
	Period                 string  `json:"period"`
}

type WakelockInfo struct {
	StuckBgWakelockRate    float64 `json:"stuck_bg_wakelock_rate"`
	StuckBgWakelockRate7d  float64 `json:"stuck_bg_wakelock_rate_7d,omitempty"`
	StuckBgWakelockRate28d float64 `json:"stuck_bg_wakelock_rate_28d,omitempty"`
	DistinctUsers          float64 `json:"distinct_users,omitempty"`
	Period                 string  `json:"period"`
}

type MemoryInfo struct {
	UserPerceivedLmkRate    float64 `json:"user_perceived_lmk_rate"`
	UserPerceivedLmkRate7d  float64 `json:"user_perceived_lmk_rate_7d,omitempty"`
	UserPerceivedLmkRate28d float64 `json:"user_perceived_lmk_rate_28d,omitempty"`
	DistinctUsers           float64 `json:"distinct_users,omitempty"`
	Period                  string  `json:"period"`
}

type ErrorInfo struct {
	ErrorReportCount float64 `json:"error_report_count"`
	DistinctUsers    float64 `json:"distinct_users,omitempty"`
	Period           string  `json:"period"`
}

func runCrashes(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	metrics, err := queryCrashMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(CrashRateInfo{
		CrashRate:     metrics["crashRate"],
		CrashRate7d:   metrics["crashRate7dUserWeighted"],
		CrashRate28d:  metrics["crashRate28dUserWeighted"],
		DistinctUsers: metrics["distinctUsers"],
		Period:        periodLabel(),
	})
}

func runANR(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	metrics, err := queryANRMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(ANRRateInfo{
		ANRRate:              metrics["anrRate"],
		ANRRate7d:            metrics["anrRate7dUserWeighted"],
		ANRRate28d:           metrics["anrRate28dUserWeighted"],
		UserPerceivedANRRate: metrics["userPerceivedAnrRate"],
		DistinctUsers:        metrics["distinctUsers"],
		Period:               periodLabel(),
	})
}

func runOverview(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	crashMetrics, err := queryCrashMetrics(client, ctx)
	if err != nil {
		return err
	}
	anrMetrics, err := queryANRMetrics(client, ctx)
	if err != nil {
		return err
	}
	slowStartMetrics, err := querySlowStartMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(VitalsOverview{
		PackageName:   cli.GetPackageName(),
		CrashRate:     crashMetrics["crashRate"],
		ANRRate:       anrMetrics["anrRate"],
		SlowStartRate: slowStartMetrics["slowStartRate"],
		Period:        periodLabel(),
	})
}

func runSlowStart(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	metrics, err := querySlowStartMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(SlowStartInfo{
		SlowStartRate:    metrics["slowStartRate"],
		SlowStartRate7d:  metrics["slowStartRate7dUserWeighted"],
		SlowStartRate28d: metrics["slowStartRate28dUserWeighted"],
		DistinctUsers:    metrics["distinctUsers"],
		Period:           periodLabel(),
	})
}

func runSlowRendering(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	metrics, err := querySlowRenderingMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(SlowRenderingInfo{
		SlowRenderingRate20Fps:    metrics["slowRenderingRate20Fps"],
		SlowRenderingRate20Fps7d:  metrics["slowRenderingRate20Fps7dUserWeighted"],
		SlowRenderingRate20Fps28d: metrics["slowRenderingRate20Fps28dUserWeighted"],
		SlowRenderingRate30Fps:    metrics["slowRenderingRate30Fps"],
		SlowRenderingRate30Fps7d:  metrics["slowRenderingRate30Fps7dUserWeighted"],
		SlowRenderingRate30Fps28d: metrics["slowRenderingRate30Fps28dUserWeighted"],
		DistinctUsers:             metrics["distinctUsers"],
		Period:                    periodLabel(),
	})
}

func runWakeups(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	metrics, err := queryWakeupMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(WakeupInfo{
		ExcessiveWakeupRate:    metrics["excessiveWakeupRate"],
		ExcessiveWakeupRate7d:  metrics["excessiveWakeupRate7dUserWeighted"],
		ExcessiveWakeupRate28d: metrics["excessiveWakeupRate28dUserWeighted"],
		DistinctUsers:          metrics["distinctUsers"],
		Period:                 periodLabel(),
	})
}

func runWakelocks(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	metrics, err := queryWakelockMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(WakelockInfo{
		StuckBgWakelockRate:    metrics["stuckBgWakelockRate"],
		StuckBgWakelockRate7d:  metrics["stuckBgWakelockRate7dUserWeighted"],
		StuckBgWakelockRate28d: metrics["stuckBgWakelockRate28dUserWeighted"],
		DistinctUsers:          metrics["distinctUsers"],
		Period:                 periodLabel(),
	})
}

func runMemory(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	metrics, err := queryMemoryMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(MemoryInfo{
		UserPerceivedLmkRate:    metrics["userPerceivedLmkRate"],
		UserPerceivedLmkRate7d:  metrics["userPerceivedLmkRate7dUserWeighted"],
		UserPerceivedLmkRate28d: metrics["userPerceivedLmkRate28dUserWeighted"],
		DistinctUsers:           metrics["distinctUsers"],
		Period:                  periodLabel(),
	})
}

func runErrors(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	metrics, err := queryErrorMetrics(client, ctx)
	if err != nil {
		return err
	}

	return output.Print(ErrorInfo{
		ErrorReportCount: metrics["errorReportCount"],
		DistinctUsers:    metrics["distinctUsers"],
		Period:           periodLabel(),
	})
}

func runErrorIssues(cmd *cobra.Command, args []string) error {
	client, ctx, cancel, err := reportingContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	resp, err := client.Vitals().Errors.Issues.Search(client.AppName()).Context(ctx).Do()
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

func reportingContext(cmd *cobra.Command) (*api.ReportingClient, context.Context, context.CancelFunc, error) {
	if err := cli.RequirePackage(cmd); err != nil {
		return nil, nil, nil, err
	}

	client, err := api.NewReportingClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, cancel := client.Context()
	return client, ctx, cancel, nil
}

func queryCrashMetrics(client *api.ReportingClient, ctx context.Context) (map[string]float64, error) {
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		Metrics:      []string{"crashRate", "crashRate7dUserWeighted", "crashRate28dUserWeighted", "distinctUsers"},
		TimelineSpec: timelineSpec(),
	}

	resp, err := client.Vitals().Crashrate.Query(
		fmt.Sprintf("%s/crashRateMetricSet", client.AppName()),
		req,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query crash rate metrics: %w", err)
	}

	return firstRowMetrics(resp.Rows)
}

func queryANRMetrics(client *api.ReportingClient, ctx context.Context) (map[string]float64, error) {
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		Metrics:      []string{"anrRate", "anrRate7dUserWeighted", "anrRate28dUserWeighted", "userPerceivedAnrRate", "distinctUsers"},
		TimelineSpec: timelineSpec(),
	}

	resp, err := client.Vitals().Anrrate.Query(
		fmt.Sprintf("%s/anrRateMetricSet", client.AppName()),
		req,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query ANR metrics: %w", err)
	}

	return firstRowMetrics(resp.Rows)
}

func querySlowStartMetrics(client *api.ReportingClient, ctx context.Context) (map[string]float64, error) {
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest{
		Metrics:      []string{"slowStartRate", "slowStartRate7dUserWeighted", "slowStartRate28dUserWeighted", "distinctUsers"},
		TimelineSpec: timelineSpec(),
	}

	resp, err := client.Vitals().Slowstartrate.Query(
		fmt.Sprintf("%s/slowStartRateMetricSet", client.AppName()),
		req,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query slow start metrics: %w", err)
	}

	return firstRowMetrics(resp.Rows)
}

func querySlowRenderingMetrics(client *api.ReportingClient, ctx context.Context) (map[string]float64, error) {
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest{
		Metrics: []string{
			"slowRenderingRate20Fps",
			"slowRenderingRate20Fps7dUserWeighted",
			"slowRenderingRate20Fps28dUserWeighted",
			"slowRenderingRate30Fps",
			"slowRenderingRate30Fps7dUserWeighted",
			"slowRenderingRate30Fps28dUserWeighted",
			"distinctUsers",
		},
		TimelineSpec: timelineSpec(),
	}

	resp, err := client.Vitals().Slowrenderingrate.Query(
		fmt.Sprintf("%s/slowRenderingRateMetricSet", client.AppName()),
		req,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query slow rendering metrics: %w", err)
	}

	return firstRowMetrics(resp.Rows)
}

func queryWakeupMetrics(client *api.ReportingClient, ctx context.Context) (map[string]float64, error) {
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest{
		Metrics:      []string{"excessiveWakeupRate", "excessiveWakeupRate7dUserWeighted", "excessiveWakeupRate28dUserWeighted", "distinctUsers"},
		TimelineSpec: timelineSpec(),
	}

	resp, err := client.Vitals().Excessivewakeuprate.Query(
		fmt.Sprintf("%s/excessiveWakeupRateMetricSet", client.AppName()),
		req,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query excessive wakeup metrics: %w", err)
	}

	return firstRowMetrics(resp.Rows)
}

func queryWakelockMetrics(client *api.ReportingClient, ctx context.Context) (map[string]float64, error) {
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest{
		Metrics:      []string{"stuckBgWakelockRate", "stuckBgWakelockRate7dUserWeighted", "stuckBgWakelockRate28dUserWeighted", "distinctUsers"},
		TimelineSpec: timelineSpec(),
	}

	resp, err := client.Vitals().Stuckbackgroundwakelockrate.Query(
		fmt.Sprintf("%s/stuckBackgroundWakelockRateMetricSet", client.AppName()),
		req,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query stuck wakelock metrics: %w", err)
	}

	return firstRowMetrics(resp.Rows)
}

func queryMemoryMetrics(client *api.ReportingClient, ctx context.Context) (map[string]float64, error) {
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryLmkRateMetricSetRequest{
		Metrics:      []string{"userPerceivedLmkRate", "userPerceivedLmkRate7dUserWeighted", "userPerceivedLmkRate28dUserWeighted", "distinctUsers"},
		TimelineSpec: timelineSpec(),
	}

	resp, err := client.Vitals().Lmkrate.Query(
		fmt.Sprintf("%s/lmkRateMetricSet", client.AppName()),
		req,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query LMK metrics: %w", err)
	}

	return firstRowMetrics(resp.Rows)
}

func queryErrorMetrics(client *api.ReportingClient, ctx context.Context) (map[string]float64, error) {
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		Metrics:      []string{"errorReportCount", "distinctUsers"},
		TimelineSpec: timelineSpec(),
	}

	resp, err := client.Vitals().Errors.Counts.Query(
		fmt.Sprintf("%s/errorCountMetricSet", client.AppName()),
		req,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query error count metrics: %w", err)
	}

	return firstRowMetrics(resp.Rows)
}

func timelineSpec() *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec {
	if days < 1 {
		days = 1
	}

	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		location = time.FixedZone("America/Los_Angeles", -8*60*60)
	}

	end := time.Now().In(location).Truncate(24 * time.Hour).Add(24 * time.Hour)
	start := end.AddDate(0, 0, -days)

	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "FULL_RANGE",
		StartTime:         dateTime(start),
		EndTime:           dateTime(end),
	}
}

func dateTime(t time.Time) *playdeveloperreporting.GoogleTypeDateTime {
	return &playdeveloperreporting.GoogleTypeDateTime{
		Year:  int64(t.Year()),
		Month: int64(t.Month()),
		Day:   int64(t.Day()),
		TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
			Id: "America/Los_Angeles",
		},
	}
}

func firstRowMetrics(rows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow) (map[string]float64, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("no vitals data returned for %s", periodLabel())
	}

	result := make(map[string]float64)
	for _, metric := range rows[0].Metrics {
		if metric == nil || metric.Metric == "" || metric.DecimalValue == nil {
			continue
		}

		value, err := strconv.ParseFloat(metric.DecimalValue.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse metric %q value %q: %w", metric.Metric, metric.DecimalValue.Value, err)
		}
		result[metric.Metric] = value
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no metric values returned for %s", periodLabel())
	}

	return result, nil
}

func periodLabel() string {
	return fmt.Sprintf("%d days", days)
}
