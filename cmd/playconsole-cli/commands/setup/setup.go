package setup

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard for Google Play Console CLI",
	Long: `Guide you through setting up the CLI with your Google Cloud project.

This wizard will help you:
  1. Create a Google Cloud Service Account
  2. Enable the Google Play Android Developer API
  3. Download and configure your credentials
  4. Link the service account in Play Console

Use --auto to automate steps 1-4 using the gcloud CLI.`,
	Run: runSetup,
}

func init() {
	SetupCmd.Flags().Bool("auto", false, "Automate setup using gcloud CLI (requires gcloud installed and authenticated)")
}

type step struct {
	title       string
	description string
	url         string
	action      string
}

var setupSteps = []step{
	{
		title:       "Create Service Account",
		description: "Create a new service account in Google Cloud Console",
		url:         "https://console.cloud.google.com/iam-admin/serviceaccounts/create",
		action:      "Name it 'play-console-cli' and click CREATE AND CONTINUE",
	},
	{
		title:       "Enable Google Play Developer API",
		description: "Enable the API that allows managing your Play Console",
		url:         "https://console.cloud.google.com/apis/library/androidpublisher.googleapis.com",
		action:      "Click the blue ENABLE button",
	},
	{
		title:       "Create Service Account Key",
		description: "Download JSON credentials for your service account",
		url:         "https://console.cloud.google.com/iam-admin/serviceaccounts",
		action:      "Click your service account → Keys tab → Add Key → JSON → Download",
	},
	{
		title:       "Link in Play Console",
		description: "Grant API access to your service account",
		url:         "https://play.google.com/console/developers/api-access",
		action:      "Link your Cloud project, find the service account, click 'Grant access'",
	},
}

func runSetup(cmd *cobra.Command, args []string) {
	autoMode, _ := cmd.Flags().GetBool("auto")
	if autoMode {
		runAutoSetup(cmd)
		return
	}

	clearScreen()
	printHeader()

	reader := bufio.NewReader(os.Stdin)

	for i, s := range setupSteps {
		printStep(i+1, len(setupSteps), s)

		fmt.Println("\nOptions:")
		fmt.Println("  [o] Open this page in browser")
		fmt.Println("  [c] Copy URL to clipboard")
		fmt.Println("  [s] Skip (already done)")
		fmt.Println("  [q] Quit setup")
		fmt.Print("\nChoice [o/c/s/q]: ")

		input := readInput(reader)

		switch input {
		case "o", "open", "":
			if err := openBrowser(s.url); err != nil {
				fmt.Printf("Warning: could not open browser: %v\n", err)
			}
			waitForUser(reader, "Press ENTER when you've completed this step...")
		case "c", "copy":
			if err := copyToClipboard(s.url); err != nil {
				fmt.Printf("Warning: could not copy to clipboard: %v\n", err)
			} else {
				fmt.Println("✓ URL copied to clipboard")
			}
			waitForUser(reader, "Press ENTER when you've completed this step...")
		case "s", "skip":
			fmt.Println("→ Skipping...")
		case "q", "quit":
			fmt.Println("\nSetup cancelled. Run 'gpc setup' to continue later.")
			return
		}

		clearScreen()
	}

	// Final step: configure credentials
	printFinalStep(reader)
}

func printHeader() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           GOOGLE PLAY CONSOLE CLI - SETUP WIZARD                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("This wizard will guide you through the one-time setup process.")
	fmt.Println("Each step will open ONE page - no loops, no spam.")
	fmt.Println()
	waitForUserSimple()
}

func printStep(current, total int, s step) {
	fmt.Printf("\n┌─ STEP %d of %d ─────────────────────────────────────────────────────┐\n", current, total)
	fmt.Printf("│  %s\n", s.title)
	fmt.Printf("├────────────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│  %s\n", s.description)
	fmt.Printf("│\n")
	fmt.Printf("│  ACTION: %s\n", s.action)
	fmt.Printf("│\n")
	fmt.Printf("│  URL: %s\n", s.url)
	fmt.Printf("└────────────────────────────────────────────────────────────────────┘\n")
}

func printFinalStep(reader *bufio.Reader) {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    CONFIGURE CREDENTIALS                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Now let's configure your downloaded credentials.")
	fmt.Println()

	// Check for existing credentials in Downloads
	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")

	jsonFiles := findJSONFiles(downloadsDir)

	var credPath string

	if len(jsonFiles) > 0 {
		fmt.Println("Found potential credential files in Downloads:")
		for i, f := range jsonFiles {
			fmt.Printf("  [%d] %s\n", i+1, filepath.Base(f))
		}
		fmt.Println("  [m] Enter path manually")
		fmt.Print("\nSelect file number or 'm': ")

		input := readInput(reader)
		if input == "m" {
			credPath = promptForPath(reader)
		} else {
			var idx int
			if _, err := fmt.Sscanf(input, "%d", &idx); err != nil {
				idx = 0
			}
			if idx >= 1 && idx <= len(jsonFiles) {
				credPath = jsonFiles[idx-1]
			} else {
				credPath = promptForPath(reader)
			}
		}
	} else {
		credPath = promptForPath(reader)
	}

	if credPath == "" {
		fmt.Println("\nNo credentials configured. Run 'gpc auth login --credentials <path>' later.")
		return
	}

	// Create config directory
	configDir := filepath.Join(homeDir, ".config", "gpc")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		return
	}

	// Copy credentials
	destPath := filepath.Join(configDir, "service-account.json")

	if err := copyFile(credPath, destPath); err != nil {
		fmt.Printf("Error copying credentials: %v\n", err)
		fmt.Println("You can manually copy your credentials to:", destPath)
		return
	}

	// Set permissions
	if err := os.Chmod(destPath, 0600); err != nil {
		fmt.Printf("Warning: could not set credentials permissions: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✓ Credentials saved to:", destPath)
	fmt.Println()

	// Ask for default package
	fmt.Print("Enter your app's package name (or press ENTER to skip): ")
	packageName := readInput(reader)

	// Run auth login
	fmt.Println()
	fmt.Println("Configuring CLI...")

	authArgs := []string{"auth", "login", "--name", "default", "--credentials", destPath}
	if packageName != "" {
		authArgs = append(authArgs, "--default-package", packageName)
	}

	// Get executable path
	execPath, _ := os.Executable()
	authCmd := exec.Command(execPath, authArgs...)
	authCmd.Stdout = os.Stdout
	authCmd.Stderr = os.Stderr

	if err := authCmd.Run(); err != nil {
		fmt.Printf("\nWarning: Could not auto-configure. Run manually:\n")
		fmt.Printf("  gpc auth login --name default --credentials %s\n", destPath)
	} else {
		fmt.Println()
		fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
		fmt.Println("║                      SETUP COMPLETE! 🎉                          ║")
		fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Println("You can now use gpc commands:")
		fmt.Println("  gpc tracks list")
		fmt.Println("  gpc bundles upload --file app.aab --track internal")
		fmt.Println("  gpc reviews list")
		fmt.Println()
		if packageName != "" {
			fmt.Printf("Default package: %s\n", packageName)
		} else {
			fmt.Println("Tip: Use --package <name> or set GPC_PACKAGE env var")
		}
	}
}

func promptForPath(reader *bufio.Reader) string {
	fmt.Print("Enter path to your service account JSON file: ")
	path := readInput(reader)

	// Expand ~ if present
	if strings.HasPrefix(path, "~") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[1:])
	}

	// Clean path to prevent traversal attacks
	path = filepath.Clean(path)

	// Ensure it's a JSON file
	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		fmt.Println("Error: File must be a .json file")
		return ""
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", path)
		return ""
	}

	return path
}

func findJSONFiles(dir string) []string {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Look for service account files
		if strings.HasSuffix(name, ".json") &&
			(strings.Contains(strings.ToLower(name), "service") ||
				strings.Contains(strings.ToLower(name), "key") ||
				strings.Contains(strings.ToLower(name), "credential")) {
			files = append(files, filepath.Join(dir, name))
		}
	}
	return files
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

func readInput(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(input))
}

func waitForUser(reader *bufio.Reader, msg string) {
	fmt.Print("\n" + msg)
	_, _ = reader.ReadString('\n')
}

func waitForUserSimple() {
	fmt.Print("Press ENTER to begin...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}

func clearScreen() {
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

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

// runGcloud executes a gcloud command and returns trimmed stdout.
func runGcloud(args ...string) (string, error) {
	cmd := exec.Command("gcloud", args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

const serviceAccountName = "play-console-cli"

// installGcloud attempts to install the gcloud CLI for the current platform.
func installGcloud() bool {
	fmt.Println()
	switch runtime.GOOS {
	case "darwin":
		// Try Homebrew first
		if _, err := exec.LookPath("brew"); err == nil {
			fmt.Println("    Installing via Homebrew...")
			cmd := exec.Command("brew", "install", "--cask", "google-cloud-sdk")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("    Homebrew install failed: %v\n", err)
				fmt.Println("    Install manually: https://cloud.google.com/sdk/install")
				return false
			}
			fmt.Println("    ✓ gcloud installed via Homebrew")
			return true
		}
		// Fall through to curl installer
		fmt.Println("    Installing via Google Cloud SDK installer...")
		cmd := exec.Command("bash", "-c", "curl -fsSL https://sdk.cloud.google.com | bash -s -- --disable-prompts")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("    Install failed: %v\n", err)
			fmt.Println("    Install manually: https://cloud.google.com/sdk/install")
			return false
		}
		fmt.Println("    ✓ gcloud installed")
		return true

	case "linux":
		fmt.Println("    Installing via Google Cloud SDK installer...")
		cmd := exec.Command("bash", "-c", "curl -fsSL https://sdk.cloud.google.com | bash -s -- --disable-prompts")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("    Install failed: %v\n", err)
			fmt.Println("    Install manually: https://cloud.google.com/sdk/install")
			return false
		}
		fmt.Println("    ✓ gcloud installed")
		return true

	default:
		fmt.Println("    Auto-install not supported on this platform.")
		fmt.Println("    Install manually: https://cloud.google.com/sdk/install")
		return false
	}
}

func runAutoSetup(cmd *cobra.Command) {
	clearScreen()

	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              MAGIC SETUP — AUTOMATED MODE                    ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// --- Check gcloud ---
	fmt.Print("==> Checking gcloud CLI... ")
	gcloudPath, lookErr := exec.LookPath("gcloud")
	if lookErr != nil {
		fmt.Println("not found")
		fmt.Println()
		fmt.Print("    Install gcloud now? [Y/n]: ")
		if input := readInput(reader); input == "n" || input == "no" {
			fmt.Println()
			fmt.Println("Run 'gpc setup' without --auto for the manual wizard.")
			return
		}
		if !installGcloud() {
			return
		}
		gcloudPath, lookErr = exec.LookPath("gcloud")
		if lookErr != nil {
			fmt.Println()
			fmt.Println("gcloud still not found. You may need to restart your terminal.")
			fmt.Println("Then try 'gpc setup --auto' again.")
			return
		}
	}
	_ = gcloudPath
	fmt.Println("✓")

	// --- Check auth ---
	fmt.Print("==> Checking gcloud auth... ")
	account, err := runGcloud("auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
	if err != nil || account == "" {
		fmt.Println("not logged in")
		fmt.Println()
		fmt.Println("    Launching gcloud login...")
		loginCmd := exec.Command("gcloud", "auth", "login")
		loginCmd.Stdin = os.Stdin
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		if err := loginCmd.Run(); err != nil {
			fmt.Println()
			fmt.Println("Login failed. Try running 'gcloud auth login' manually,")
			fmt.Println("then 'gpc setup --auto' again.")
			return
		}
		// Re-check
		account, _ = runGcloud("auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
		if account == "" {
			fmt.Println("Login did not complete. Try 'gpc setup --auto' again.")
			return
		}
	}
	fmt.Printf("✓ (%s)\n", account)
	fmt.Println()

	// --- Step 1: Project ---
	fmt.Print("==> [1/5] Google Cloud Project... ")
	project, _ := runGcloud("config", "get-value", "project")
	if project == "" || project == "(unset)" {
		fmt.Println()
		fmt.Print("    No project set. Enter your GCP project ID: ")
		project = readInput(reader)
		if project == "" {
			fmt.Println("    Project ID is required. Aborting.")
			return
		}
		if _, err := runGcloud("config", "set", "project", project); err != nil {
			fmt.Printf("    Failed to set project: %v\n", err)
			return
		}
		fmt.Printf("    Set project: %s\n", project)
	} else {
		fmt.Printf("✓ (%s)\n", project)
	}

	// --- Step 2: Enable API ---
	fmt.Print("==> [2/5] Enabling Google Play Android Developer API... ")
	enabledAPIs, _ := runGcloud("services", "list", "--enabled", "--filter=name:androidpublisher.googleapis.com", "--format=value(name)", "--project", project)
	if strings.Contains(enabledAPIs, "androidpublisher") {
		fmt.Println("✓ (already enabled)")
	} else {
		_, err = runGcloud("services", "enable", "androidpublisher.googleapis.com", "--project", project)
		if err != nil {
			fmt.Println("✗")
			fmt.Printf("    Error: %v\n", err)
			return
		}
		fmt.Println("✓")
	}

	// --- Step 3: Create service account ---
	saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccountName, project)
	fmt.Print("==> [3/5] Creating service account... ")
	existingSA, _ := runGcloud("iam", "service-accounts", "list", "--filter=email:"+saEmail, "--format=value(email)", "--project", project)
	if existingSA == saEmail {
		fmt.Println("✓ (already exists)")
	} else {
		_, err = runGcloud("iam", "service-accounts", "create", serviceAccountName,
			"--display-name", "Play Console CLI",
			"--project", project)
		if err != nil {
			fmt.Println("✗")
			fmt.Printf("    Error: %v\n", err)
			return
		}
		fmt.Println("✓")
	}
	fmt.Printf("    Email: %s\n", saEmail)

	// --- Step 4: Download key ---
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "gpc")
	keyPath := filepath.Join(configDir, "service-account.json")

	fmt.Print("==> [4/5] Downloading credentials... ")

	// Check if key file already exists
	if _, err := os.Stat(keyPath); err == nil {
		fmt.Println()
		fmt.Printf("    Credentials already exist at %s\n", keyPath)
		fmt.Print("    Overwrite? [y/N]: ")
		if input := readInput(reader); input != "y" && input != "yes" {
			fmt.Println("    Keeping existing credentials.")
			goto step5
		}
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Printf("    Error creating config directory: %v\n", err)
		return
	}
	_, err = runGcloud("iam", "service-accounts", "keys", "create", keyPath,
		"--iam-account", saEmail,
		"--project", project)
	if err != nil {
		fmt.Println("✗")
		fmt.Printf("    Error: %v\n", err)
		return
	}
	if err := os.Chmod(keyPath, 0600); err != nil {
		fmt.Printf("    Warning: could not set key permissions: %v\n", err)
	}
	fmt.Println("✓")
	fmt.Printf("    Saved to %s\n", keyPath)

step5:
	// --- Step 5: Manual — Grant access in Play Console ---
	playConsoleURL := "https://play.google.com/console/developers/api-access"
	fmt.Println()
	fmt.Println("==> [5/5] Grant access in Play Console")
	fmt.Println()
	fmt.Println("    Almost done! This is the one step that can't be automated.")
	fmt.Println("    Opening Play Console now...")
	fmt.Println()
	fmt.Println("    Just do this:")
	fmt.Println("      1. Click \"Link\" next to your Cloud project")
	fmt.Printf("      2. Find: %s\n", saEmail)
	fmt.Println("      3. Click \"Grant access\" → set permissions → \"Invite user\"")
	fmt.Println()
	if err := copyToClipboard(saEmail); err != nil {
		fmt.Printf("    Warning: could not copy service account email: %v\n", err)
	} else {
		fmt.Println("    (Service account email copied to clipboard)")
	}
	if err := openBrowser(playConsoleURL); err != nil {
		fmt.Printf("    Warning: could not open browser: %v\n", err)
	}
	waitForUser(reader, "    Press ENTER when done...")

	// --- Configure CLI profile ---
	execPath, _ := os.Executable()

	fmt.Println()
	fmt.Print("Enter your app's package name (or ENTER to skip): ")
	packageName := readInput(reader)

	fmt.Print("==> Configuring CLI profile... ")

	authArgs := []string{"auth", "login", "--name", "default", "--credentials", keyPath}
	if packageName != "" {
		authArgs = append(authArgs, "--default-package", packageName)
	}

	authCmd := exec.Command(execPath, authArgs...)
	authCmd.Stdout = nil
	authCmd.Stderr = nil

	if err := authCmd.Run(); err != nil {
		fmt.Println("✗")
		fmt.Printf("    Run manually: gpc auth login --name default --credentials %s\n", keyPath)
	} else {
		fmt.Println("✓")
	}

	// --- Done ---
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     SETUP COMPLETE!                          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("You can now use gpc commands:")
	fmt.Println("  gpc tracks list")
	fmt.Println("  gpc bundles upload --file app.aab --track internal")
	fmt.Println("  gpc reviews list")
	fmt.Println()
	if packageName != "" {
		fmt.Printf("Default package: %s\n", packageName)
	} else {
		fmt.Println("Tip: Use --package <name> or set GPC_PACKAGE env var")
	}
	fmt.Println()
	fmt.Println("Run 'gpc doctor' to verify everything is connected.")
}
