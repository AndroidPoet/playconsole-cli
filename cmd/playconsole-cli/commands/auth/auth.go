package auth

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/config"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication profiles",
	Long:  `Manage authentication profiles for Google Play Console API access.`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure authentication credentials",
	Long: `Configure authentication credentials for Google Play Console API.

You need a service account with access to the Google Play Developer API.
Create one at: https://console.cloud.google.com/iam-admin/serviceaccounts

Grant access at: https://play.google.com/console/developers/api-access`,
	RunE: runLogin,
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch to a different profile",
	RunE:  runSwitch,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured profiles",
	RunE:  runList,
}

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current profile",
	RunE:  runCurrent,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a profile",
	RunE:  runDelete,
}

var (
	profileName     string
	credentialsPath string
	credentialsB64  string
	defaultPackage  string
)

func init() {
	// Login flags
	loginCmd.Flags().StringVar(&profileName, "name", "default", "profile name")
	loginCmd.Flags().StringVar(&credentialsPath, "credentials", "", "path to service account JSON file")
	loginCmd.Flags().StringVar(&credentialsB64, "credentials-base64", "", "base64-encoded service account JSON")
	loginCmd.Flags().StringVar(&defaultPackage, "default-package", "", "default package name for this profile")

	// Switch flags
	switchCmd.Flags().StringVar(&profileName, "name", "", "profile name to switch to")
	switchCmd.MarkFlagRequired("name")

	// Delete flags
	deleteCmd.Flags().StringVar(&profileName, "name", "", "profile name to delete")
	deleteCmd.MarkFlagRequired("name")
	deleteCmd.Flags().Bool("confirm", false, "confirm deletion")

	AuthCmd.AddCommand(loginCmd)
	AuthCmd.AddCommand(switchCmd)
	AuthCmd.AddCommand(listCmd)
	AuthCmd.AddCommand(currentCmd)
	AuthCmd.AddCommand(deleteCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Validate credentials source
	if credentialsPath == "" && credentialsB64 == "" {
		return fmt.Errorf("either --credentials or --credentials-base64 is required")
	}

	if credentialsPath != "" && credentialsB64 != "" {
		return fmt.Errorf("only one of --credentials or --credentials-base64 should be specified")
	}

	// Validate credentials file exists
	if credentialsPath != "" {
		absPath, err := filepath.Abs(credentialsPath)
		if err != nil {
			return fmt.Errorf("invalid credentials path: %w", err)
		}
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("credentials file not found: %s", absPath)
		}
		credentialsPath = absPath
	}

	// Validate base64 credentials
	if credentialsB64 != "" {
		if _, err := base64.StdEncoding.DecodeString(credentialsB64); err != nil {
			return fmt.Errorf("invalid base64 credentials: %w", err)
		}
	}

	// Create profile
	profile := config.Profile{
		Name:            profileName,
		CredentialsPath: credentialsPath,
		CredentialsB64:  credentialsB64,
		DefaultPackage:  defaultPackage,
	}

	// Save to config
	config.SetProfile(profile)
	if err := config.Save(); err != nil {
		return err
	}

	output.PrintSuccess("Profile '%s' configured successfully", profileName)

	// Show configured profile
	return output.Print(map[string]interface{}{
		"profile":         profileName,
		"credentials":     credentialsPath,
		"default_package": defaultPackage,
	})
}

func runSwitch(cmd *cobra.Command, args []string) error {
	// Check if profile exists
	profiles := config.ListProfiles()
	found := false
	for _, p := range profiles {
		if p == profileName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile '%s' not found. Use 'gpc auth list' to see available profiles", profileName)
	}

	// Update default profile
	config.SetDefaultProfile(profileName)
	if err := config.Save(); err != nil {
		return err
	}

	output.PrintSuccess("Switched to profile '%s'", profileName)
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	cfg := config.GetConfig()

	type ProfileInfo struct {
		Name      string `json:"name"`
		Default   bool   `json:"default"`
		Package   string `json:"package,omitempty"`
		CredsType string `json:"credentials_type"`
	}

	profiles := make([]ProfileInfo, 0)
	for name, p := range cfg.Profiles {
		credsType := "file"
		if p.CredentialsB64 != "" {
			credsType = "base64"
		}
		profiles = append(profiles, ProfileInfo{
			Name:      name,
			Default:   name == cfg.DefaultProfile,
			Package:   p.DefaultPackage,
			CredsType: credsType,
		})
	}

	if len(profiles) == 0 {
		output.PrintInfo("No profiles configured. Run 'gpc auth login' to add one.")
		return nil
	}

	return output.Print(profiles)
}

func runCurrent(cmd *cobra.Command, args []string) error {
	profile := config.GetProfile()
	if profile == nil {
		return fmt.Errorf("no profile configured")
	}

	credsType := "file"
	if profile.CredentialsB64 != "" {
		credsType = "base64"
	}

	return output.Print(map[string]interface{}{
		"name":             profile.Name,
		"credentials_type": credsType,
		"default_package":  profile.DefaultPackage,
	})
}

func runDelete(cmd *cobra.Command, args []string) error {
	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("use --confirm to delete profile '%s'", profileName)
	}

	config.DeleteProfile(profileName)
	if err := config.Save(); err != nil {
		return err
	}

	output.PrintSuccess("Profile '%s' deleted", profileName)
	return nil
}
