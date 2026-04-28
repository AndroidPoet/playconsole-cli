package api

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// APIEnablementError represents an error when an API needs to be enabled
type APIEnablementError struct {
	APIName       string
	APITitle      string
	ProjectID     string
	ActivationURL string
	OriginalError error
}

func (e *APIEnablementError) Error() string {
	return fmt.Sprintf("API '%s' is not enabled", e.APIName)
}

// ParseAPIEnablementError checks if an error is due to a disabled API
// and extracts the relevant information
func ParseAPIEnablementError(err error) *APIEnablementError {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check if this is a SERVICE_DISABLED error
	if !strings.Contains(errStr, "SERVICE_DISABLED") && !strings.Contains(errStr, "has not been used in project") {
		return nil
	}

	// Extract activation URL
	urlRegex := regexp.MustCompile(`https://console\.developers\.google\.com/apis/api/([^/]+)/overview\?project=(\d+)`)
	matches := urlRegex.FindStringSubmatch(errStr)

	if len(matches) < 3 {
		// Try alternate URL format
		urlRegex = regexp.MustCompile(`https://console\.cloud\.google\.com/apis/api/([^/]+)/overview\?project=(\d+)`)
		matches = urlRegex.FindStringSubmatch(errStr)
	}

	if len(matches) < 3 {
		return nil
	}

	apiName := matches[1]
	projectID := matches[2]

	// Extract API title if available
	titleRegex := regexp.MustCompile(`"serviceTitle":\s*"([^"]+)"`)
	titleMatches := titleRegex.FindStringSubmatch(errStr)
	apiTitle := apiName
	if len(titleMatches) >= 2 {
		apiTitle = titleMatches[1]
	}

	return &APIEnablementError{
		APIName:       apiName,
		APITitle:      apiTitle,
		ProjectID:     projectID,
		ActivationURL: matches[0],
		OriginalError: err,
	}
}

// HandleAPIEnablement provides an interactive flow to enable a required API
func HandleAPIEnablement(apiErr *APIEnablementError) error {
	if !isInteractiveSession() {
		return fmt.Errorf(
			"API '%s' must be enabled before retrying. Enable it here: %s",
			apiErr.APITitle,
			apiErr.ActivationURL,
		)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     API ENABLEMENT REQUIRED                       ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  API: %-58s ║\n", apiErr.APITitle)
	fmt.Printf("║  Project: %-54s ║\n", apiErr.ProjectID)
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("This API needs to be enabled in your Google Cloud project.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  [o] Open browser to enable API")
	fmt.Println("  [c] Copy URL to clipboard")
	fmt.Println("  [s] Show URL only")
	fmt.Println("  [q] Quit")
	fmt.Println()
	fmt.Print("Choose an option [o/c/s/q]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "o", "open", "":
		fmt.Println()
		fmt.Println("Opening browser...")
		if err := openBrowser(apiErr.ActivationURL); err != nil {
			fmt.Printf("Could not open browser: %v\n", err)
			fmt.Println()
			fmt.Println("Please open this URL manually:")
			fmt.Println(apiErr.ActivationURL)
		}
		return waitForEnablement()

	case "c", "copy":
		if err := copyToClipboard(apiErr.ActivationURL); err != nil {
			fmt.Printf("Could not copy to clipboard: %v\n", err)
			fmt.Println()
			fmt.Println("URL:")
			fmt.Println(apiErr.ActivationURL)
		} else {
			fmt.Println()
			fmt.Println("URL copied to clipboard!")
		}
		return waitForEnablement()

	case "s", "show":
		fmt.Println()
		fmt.Println("Enable URL:")
		fmt.Println(apiErr.ActivationURL)
		return waitForEnablement()

	case "q", "quit":
		return fmt.Errorf("API enablement cancelled by user")

	default:
		fmt.Println("Invalid option")
		return HandleAPIEnablement(apiErr)
	}
}

func isInteractiveSession() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func waitForEnablement() error {
	fmt.Println()
	fmt.Println("After enabling the API in Google Cloud Console:")
	fmt.Println("  1. Click 'ENABLE' on the API page")
	fmt.Println("  2. Wait a few seconds for it to activate")
	fmt.Println()
	fmt.Print("Press ENTER when done (or 'q' to quit): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "q" || input == "quit" {
		return fmt.Errorf("API enablement cancelled by user")
	}

	fmt.Println()
	fmt.Println("Retrying...")
	return nil
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// copyToClipboard copies text to the system clipboard
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// RequiredAPIs lists all APIs that might be needed by the CLI
var RequiredAPIs = map[string]string{
	"androidpublisher.googleapis.com":       "Google Play Android Developer API",
	"playdeveloperreporting.googleapis.com": "Google Play Developer Reporting API",
}

// CheckAndEnableAPI wraps an API call and handles enablement if needed
func CheckAndEnableAPI(operation func() error) error {
	for {
		err := operation()
		if err == nil {
			return nil
		}

		apiErr := ParseAPIEnablementError(err)
		if apiErr == nil {
			// Not an enablement error, return as-is
			return err
		}

		// Handle enablement flow
		if handleErr := HandleAPIEnablement(apiErr); handleErr != nil {
			return handleErr
		}

		// Retry the operation
	}
}
