package api

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/AndroidPoet/playconsole-cli/internal/config"
)

// ReportingClient wraps the Play Developer Reporting API client
type ReportingClient struct {
	service     *playdeveloperreporting.Service
	packageName string
	timeout     time.Duration
}

// NewReportingClient creates a new Reporting API client
func NewReportingClient(packageName string, timeout time.Duration) (*ReportingClient, error) {
	ctx := context.Background()

	// Get credentials
	creds, err := config.GetCredentials()
	if err != nil {
		return nil, err
	}

	// Create JWT config with reporting scope
	jwtConfig, err := google.JWTConfigFromJSON(creds, playdeveloperreporting.PlaydeveloperreportingScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	// Create HTTP client
	httpClient := jwtConfig.Client(ctx)

	// Add debug transport if enabled
	if config.IsDebug() {
		httpClient.Transport = &debugTransport{base: httpClient.Transport}
	}

	// Create service
	service, err := playdeveloperreporting.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Reporting API client: %w", err)
	}

	return &ReportingClient{
		service:     service,
		packageName: packageName,
		timeout:     timeout,
	}, nil
}

// Context returns a context with timeout
func (c *ReportingClient) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

// GetPackageName returns the package name
func (c *ReportingClient) GetPackageName() string {
	return c.packageName
}

// AppName returns the formatted app name for API calls
func (c *ReportingClient) AppName() string {
	return fmt.Sprintf("apps/%s", c.packageName)
}

// Vitals returns the vitals service
func (c *ReportingClient) Vitals() *playdeveloperreporting.VitalsService {
	return c.service.Vitals
}

// Anomalies returns the anomalies service
func (c *ReportingClient) Anomalies() *playdeveloperreporting.AnomaliesService {
	return c.service.Anomalies
}
