package doctor

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/config"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// DoctorCmd validates CLI setup and credentials
var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate CLI setup and credentials",
	Long: `Run diagnostic checks to verify your playconsole-cli setup.

Checks configuration, credentials, API connectivity, and permissions
to help troubleshoot common issues.`,
	RunE: runDoctor,
}

var verbose bool

func init() {
	DoctorCmd.Flags().BoolVar(&verbose, "verbose", false, "show detailed check output")
}

// CheckResult represents a single diagnostic check
type CheckResult struct {
	Check   string `json:"check"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	results := make([]CheckResult, 0, 6)

	// 1. Config file
	results = append(results, checkConfig())

	// 2. Credentials
	results = append(results, checkCredentials())

	// 3. Service account validity
	results = append(results, checkServiceAccount())

	// 4. Package name
	results = append(results, checkPackageName())

	// 5. Android Publisher API
	pkgName := cli.GetPackageName()
	if pkgName != "" {
		results = append(results, checkPublisherAPI(pkgName))
		results = append(results, checkReportingAPI(pkgName))
	} else {
		results = append(results, CheckResult{
			Check:   "Android Publisher API",
			Status:  "skip",
			Message: "no package name configured, skipping API checks",
		})
		results = append(results, CheckResult{
			Check:   "Reporting API",
			Status:  "skip",
			Message: "no package name configured, skipping API checks",
		})
	}

	// Print summary
	passed := 0
	failed := 0
	skipped := 0
	for _, r := range results {
		switch r.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		case "skip":
			skipped++
		}
	}

	if err := output.Print(results); err != nil {
		return err
	}

	output.PrintInfo("\n%d passed, %d failed, %d skipped", passed, failed, skipped)

	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}

	return nil
}

func checkConfig() CheckResult {
	err := config.Init("", "")
	if err != nil {
		return CheckResult{
			Check:   "Configuration",
			Status:  "fail",
			Message: fmt.Sprintf("config error: %v", err),
		}
	}
	return CheckResult{
		Check:   "Configuration",
		Status:  "pass",
		Message: "config loaded successfully",
	}
}

func checkCredentials() CheckResult {
	creds, err := config.GetCredentials()
	if err != nil {
		return CheckResult{
			Check:   "Credentials",
			Status:  "fail",
			Message: fmt.Sprintf("credentials not found: %v", err),
		}
	}
	if len(creds) == 0 {
		return CheckResult{
			Check:   "Credentials",
			Status:  "fail",
			Message: "credentials file is empty",
		}
	}
	return CheckResult{
		Check:   "Credentials",
		Status:  "pass",
		Message: "credentials available",
	}
}

func checkServiceAccount() CheckResult {
	creds, err := config.GetCredentials()
	if err != nil {
		return CheckResult{
			Check:   "Service Account",
			Status:  "skip",
			Message: "no credentials to validate",
		}
	}

	var sa struct {
		Type         string `json:"type"`
		ProjectID    string `json:"project_id"`
		ClientEmail  string `json:"client_email"`
		PrivateKeyID string `json:"private_key_id"`
	}
	if err := json.Unmarshal(creds, &sa); err != nil {
		return CheckResult{
			Check:   "Service Account",
			Status:  "fail",
			Message: fmt.Sprintf("invalid JSON: %v", err),
		}
	}

	if sa.Type != "service_account" {
		return CheckResult{
			Check:   "Service Account",
			Status:  "fail",
			Message: fmt.Sprintf("unexpected type '%s', expected 'service_account'", sa.Type),
		}
	}

	msg := fmt.Sprintf("project=%s email=%s", sa.ProjectID, sa.ClientEmail)
	return CheckResult{
		Check:   "Service Account",
		Status:  "pass",
		Message: msg,
	}
}

func checkPackageName() CheckResult {
	pkg := cli.GetPackageName()
	if pkg == "" {
		return CheckResult{
			Check:   "Package Name",
			Status:  "warn",
			Message: "no package name set (use --package or GPC_PACKAGE env)",
		}
	}
	return CheckResult{
		Check:   "Package Name",
		Status:  "pass",
		Message: pkg,
	}
}

func checkPublisherAPI(packageName string) CheckResult {
	client, err := api.NewClient(packageName, 30*time.Second)
	if err != nil {
		return CheckResult{
			Check:   "Android Publisher API",
			Status:  "fail",
			Message: fmt.Sprintf("client creation failed: %v", err),
		}
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Try to create and immediately delete an edit as a connectivity test
	edit, err := client.Edits().Insert(packageName, nil).Context(ctx).Do()
	if err != nil {
		return CheckResult{
			Check:   "Android Publisher API",
			Status:  "fail",
			Message: fmt.Sprintf("API call failed: %v", err),
		}
	}

	// Clean up the test edit
	_ = client.Edits().Delete(packageName, edit.Id).Context(ctx).Do()

	return CheckResult{
		Check:   "Android Publisher API",
		Status:  "pass",
		Message: "API is reachable and authenticated",
	}
}

func checkReportingAPI(packageName string) CheckResult {
	client, err := api.NewReportingClient(packageName, 30*time.Second)
	if err != nil {
		return CheckResult{
			Check:   "Reporting API",
			Status:  "fail",
			Message: fmt.Sprintf("client creation failed: %v", err),
		}
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Try a lightweight API call
	appName := client.AppName()
	crashRateName := fmt.Sprintf("%s/crashRateMetricSet", appName)
	_, err = client.Vitals().Crashrate.Get(crashRateName).Context(ctx).Do()
	if err != nil {
		return CheckResult{
			Check:   "Reporting API",
			Status:  "fail",
			Message: fmt.Sprintf("API call failed: %v", err),
		}
	}

	return CheckResult{
		Check:   "Reporting API",
		Status:  "pass",
		Message: "API is reachable and authenticated",
	}
}
