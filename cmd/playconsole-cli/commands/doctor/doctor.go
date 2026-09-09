package doctor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
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
to help troubleshoot common issues. Honors the global --config, --profile
and --package flags, so you can verify exactly the setup a command would use.`,
	RunE: runDoctor,
}

var verbose bool

func init() {
	DoctorCmd.Flags().BoolVar(&verbose, "verbose", false, "include details (paths, identities, latency) in each check")
}

// CheckResult represents a single diagnostic check
type CheckResult struct {
	Check   string `json:"check"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
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

	// 5/6. API connectivity
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

	// Strip details unless asked for them
	if !verbose {
		for i := range results {
			results[i].Detail = ""
		}
	}

	// Print summary
	passed, failed, warned, skipped := 0, 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		case "warn":
			warned++
		case "skip":
			skipped++
		}
	}

	if err := output.Print(results); err != nil {
		return err
	}

	output.PrintInfo("%d passed, %d failed, %d warnings, %d skipped", passed, failed, warned, skipped)

	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}

	return nil
}

// checkConfig reports on the configuration that was already loaded for this
// invocation (respecting --config / --profile), rather than re-loading defaults.
func checkConfig() CheckResult {
	path := config.GetConfigPath()
	profile := config.GetProfile()
	profileName := ""
	if profile != nil {
		profileName = profile.Name
	}
	detail := fmt.Sprintf("config=%s profile=%s", path, profileName)

	if _, err := os.Stat(path); err != nil {
		return CheckResult{
			Check:   "Configuration",
			Status:  "warn",
			Message: fmt.Sprintf("no config file at %s (using flags/env only)", path),
			Detail:  detail,
		}
	}

	cfg := config.GetConfig()
	if cfg != nil {
		if _, ok := cfg.Profiles[profileName]; !ok && profileName != "" {
			return CheckResult{
				Check:   "Configuration",
				Status:  "warn",
				Message: fmt.Sprintf("profile '%s' not found in %s (using env credentials if set)", profileName, path),
				Detail:  detail,
			}
		}
	}

	return CheckResult{
		Check:   "Configuration",
		Status:  "pass",
		Message: fmt.Sprintf("config loaded, profile '%s'", profileName),
		Detail:  detail,
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

	source := "unknown"
	if p := config.GetProfile(); p != nil {
		switch {
		case p.CredentialsB64 != "":
			source = fmt.Sprintf("base64 (%d bytes decoded)", len(creds))
		case p.CredentialsPath != "":
			source = "file " + p.CredentialsPath
		}
	}
	return CheckResult{
		Check:   "Credentials",
		Status:  "pass",
		Message: "credentials available",
		Detail:  source,
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
		PrivateKey   string `json:"private_key"`
	}
	if err := json.Unmarshal(creds, &sa); err != nil {
		// A common mistake is base64-encoding twice; give a targeted hint.
		if _, b64Err := base64.StdEncoding.DecodeString(string(creds)); b64Err == nil {
			return CheckResult{
				Check:   "Service Account",
				Status:  "fail",
				Message: "credentials are base64 text, not JSON (encoded twice?)",
			}
		}
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
	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return CheckResult{
			Check:   "Service Account",
			Status:  "fail",
			Message: "service account JSON is missing client_email or private_key",
		}
	}

	return CheckResult{
		Check:   "Service Account",
		Status:  "pass",
		Message: fmt.Sprintf("project=%s email=%s", sa.ProjectID, sa.ClientEmail),
		Detail:  "key_id=" + sa.PrivateKeyID,
	}
}

func checkPackageName() CheckResult {
	pkg := cli.GetPackageName()
	if pkg == "" {
		return CheckResult{
			Check:   "Package Name",
			Status:  "warn",
			Message: "no package name set (use --package, GPC_PACKAGE, or a profile default)",
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

	// Creating (and discarding) an edit exercises auth and package access.
	started := time.Now()
	edit, err := client.CreateEdit()
	if err != nil {
		if apiErr := api.ParseAPIEnablementError(err); apiErr != nil {
			return CheckResult{
				Check:   "Android Publisher API",
				Status:  "fail",
				Message: fmt.Sprintf("API not enabled in project %s; enable it at %s", apiErr.ProjectID, apiErr.ActivationURL),
			}
		}
		return CheckResult{
			Check:   "Android Publisher API",
			Status:  "fail",
			Message: fmt.Sprintf("API call failed: %v", err),
		}
	}
	edit.Close()

	return CheckResult{
		Check:   "Android Publisher API",
		Status:  "pass",
		Message: "API is reachable and authenticated",
		Detail:  fmt.Sprintf("edit create+delete round-trip %s", time.Since(started).Round(time.Millisecond)),
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

	started := time.Now()
	crashRateName := fmt.Sprintf("%s/crashRateMetricSet", client.AppName())
	_, err = client.Vitals().Crashrate.Get(crashRateName).Context(ctx).Do()
	if err != nil {
		if apiErr := api.ParseAPIEnablementError(err); apiErr != nil {
			return CheckResult{
				Check:   "Reporting API",
				Status:  "fail",
				Message: fmt.Sprintf("API not enabled in project %s; enable it at %s", apiErr.ProjectID, apiErr.ActivationURL),
			}
		}
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
		Detail:  fmt.Sprintf("crash rate metric-set lookup %s", time.Since(started).Round(time.Millisecond)),
	}
}
