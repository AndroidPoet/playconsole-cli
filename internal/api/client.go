package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"

	"github.com/AndroidPoet/playconsole-cli/internal/config"
)

// Client wraps the Android Publisher API client
type Client struct {
	service     *androidpublisher.Service
	packageName string
	timeout     time.Duration
	debug       bool
}

// debugTransport wraps http.RoundTripper to log requests
type debugTransport struct {
	base http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("DEBUG: %s %s\n", req.Method, req.URL)
	return t.base.RoundTrip(req)
}

// NewClient creates a new API client
func NewClient(packageName string, timeout time.Duration) (*Client, error) {
	ctx := context.Background()

	// Get credentials
	creds, err := config.GetCredentials()
	if err != nil {
		return nil, err
	}

	// Create JWT config
	jwtConfig, err := google.JWTConfigFromJSON(creds, androidpublisher.AndroidpublisherScope)
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
	service, err := androidpublisher.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &Client{
		service:     service,
		packageName: packageName,
		timeout:     timeout,
		debug:       config.IsDebug(),
	}, nil
}

// GetService returns the underlying Android Publisher service
func (c *Client) GetService() *androidpublisher.Service {
	return c.service
}

// GetPackageName returns the package name
func (c *Client) GetPackageName() string {
	return c.packageName
}

// Context returns a context with timeout
func (c *Client) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

// Apps returns the applications service
func (c *Client) Apps() *androidpublisher.ApplicationsService {
	return c.service.Applications
}

// Edits returns the edits service
func (c *Client) Edits() *androidpublisher.EditsService {
	return c.service.Edits
}

// InAppProducts returns the in-app products service
func (c *Client) InAppProducts() *androidpublisher.InappproductsService {
	return c.service.Inappproducts
}

// Reviews returns the reviews service
func (c *Client) Reviews() *androidpublisher.ReviewsService {
	return c.service.Reviews
}

// Purchases returns the purchases service
func (c *Client) Purchases() *androidpublisher.PurchasesService {
	return c.service.Purchases
}

// Monetization returns the monetization service
func (c *Client) Monetization() *androidpublisher.MonetizationService {
	return c.service.Monetization
}

// Users returns the users service
func (c *Client) Users() *androidpublisher.UsersService {
	return c.service.Users
}

// Grants returns the grants service
func (c *Client) Grants() *androidpublisher.GrantsService {
	return c.service.Grants
}

// Edit represents an active edit session
type Edit struct {
	client  *Client
	editID  string
	ctx     context.Context
	cancel  context.CancelFunc
}

// CreateEdit creates a new edit session
func (c *Client) CreateEdit() (*Edit, error) {
	ctx, cancel := c.Context()

	edit, err := c.service.Edits.Insert(c.packageName, nil).Context(ctx).Do()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create edit: %w", err)
	}

	return &Edit{
		client: c,
		editID: edit.Id,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// GetEdit returns an existing edit by ID
func (c *Client) GetEdit(editID string) (*Edit, error) {
	ctx, cancel := c.Context()

	edit, err := c.service.Edits.Get(c.packageName, editID).Context(ctx).Do()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get edit: %w", err)
	}

	return &Edit{
		client: c,
		editID: edit.Id,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// ID returns the edit ID
func (e *Edit) ID() string {
	return e.editID
}

// Context returns the edit context
func (e *Edit) Context() context.Context {
	return e.ctx
}

// Validate validates the edit
func (e *Edit) Validate() error {
	_, err := e.client.service.Edits.Validate(e.client.packageName, e.editID).Context(e.ctx).Do()
	if err != nil {
		return fmt.Errorf("edit validation failed: %w", err)
	}
	return nil
}

// Commit commits the edit
func (e *Edit) Commit() error {
	_, err := e.client.service.Edits.Commit(e.client.packageName, e.editID).Context(e.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to commit edit: %w", err)
	}
	return nil
}

// Delete deletes the edit without committing
func (e *Edit) Delete() error {
	err := e.client.service.Edits.Delete(e.client.packageName, e.editID).Context(e.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete edit: %w", err)
	}
	return nil
}

// Close releases resources
func (e *Edit) Close() {
	e.cancel()
}

// Tracks returns the tracks service for this edit
func (e *Edit) Tracks() *androidpublisher.EditsTracksService {
	return e.client.service.Edits.Tracks
}

// Bundles returns the bundles service for this edit
func (e *Edit) Bundles() *androidpublisher.EditsBundlesService {
	return e.client.service.Edits.Bundles
}

// APKs returns the APKs service for this edit
func (e *Edit) APKs() *androidpublisher.EditsApksService {
	return e.client.service.Edits.Apks
}

// Listings returns the listings service for this edit
func (e *Edit) Listings() *androidpublisher.EditsListingsService {
	return e.client.service.Edits.Listings
}

// Images returns the images service for this edit
func (e *Edit) Images() *androidpublisher.EditsImagesService {
	return e.client.service.Edits.Images
}

// Details returns the details service for this edit
func (e *Edit) Details() *androidpublisher.EditsDetailsService {
	return e.client.service.Edits.Details
}

// Testers returns the testers service for this edit
func (e *Edit) Testers() *androidpublisher.EditsTestersService {
	return e.client.service.Edits.Testers
}

// DeobfuscationFiles returns the deobfuscation files service for this edit
func (e *Edit) DeobfuscationFiles() *androidpublisher.EditsDeobfuscationfilesService {
	return e.client.service.Edits.Deobfuscationfiles
}

// ExpansionFiles returns the expansion files service for this edit
func (e *Edit) ExpansionFiles() *androidpublisher.EditsExpansionfilesService {
	return e.client.service.Edits.Expansionfiles
}
