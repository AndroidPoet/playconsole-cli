package users

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var UsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage user access and permissions",
	Long: `Manage user access to your Google Play Console account.

This allows you to grant, modify, and revoke access for team members
to specific apps or the entire developer account.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE:  runList,
}

var grantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant app access to a user",
	RunE:  runGrant,
}

var revokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke app access from a user",
	RunE:  runRevoke,
}

var (
	email       string
	role        string
	accessLevel string
)

// Available access levels
var validAccessLevels = []string{
	"accessLevel/admin",
	"accessLevel/releaseManager",
	"accessLevel/appOwner",
}

func init() {
	// Grant flags
	grantCmd.Flags().StringVar(&email, "email", "", "user email")
	grantCmd.Flags().StringVar(&role, "role", "releaseManager", "role: admin, releaseManager, appOwner")
	grantCmd.MarkFlagRequired("email")

	// Revoke flags
	revokeCmd.Flags().StringVar(&email, "email", "", "user email")
	revokeCmd.Flags().Bool("confirm", false, "confirm revocation")
	revokeCmd.MarkFlagRequired("email")

	UsersCmd.AddCommand(listCmd)
	UsersCmd.AddCommand(grantCmd)
	UsersCmd.AddCommand(revokeCmd)
}

// UserInfo represents user information
type UserInfo struct {
	Email           string   `json:"email"`
	Name            string   `json:"name,omitempty"`
	DeveloperAccess string   `json:"developer_access,omitempty"`
	AppAccess       []string `json:"app_access,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := api.NewClient("", 60*time.Second) // No package needed for listing users
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Note: The Users API requires the developer ID, which is obtained from the parent
	// For this implementation, we'll show how the API would be called
	users, err := client.Users().List("developers/-").Context(ctx).Do()
	if err != nil {
		return err
	}

	result := make([]UserInfo, 0, len(users.Users))
	for _, u := range users.Users {
		info := UserInfo{
			Email: u.Email,
			Name:  u.Name,
		}

		// Developer access info would be populated from grants
		info.DeveloperAccess = "configured"

		result = append(result, info)
	}

	if len(result) == 0 {
		output.PrintInfo("No users found")
		return nil
	}

	return output.Print(result)
}

func runGrant(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	// Map role flag to access level
	accessLevel := fmt.Sprintf("accessLevel/%s", role)

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would grant '%s' access to %s for package %s", role, email, cli.GetPackageName())
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Create a grant for the user
	grant := &androidpublisher.Grant{
		PackageName: cli.GetPackageName(),
		AppLevelPermissions: []string{
			accessLevel,
		},
	}

	// The parent path for grants is: developers/{developer_id}/users/{email}
	parent := fmt.Sprintf("developers/-/users/%s", email)

	created, err := client.Grants().Create(parent, grant).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Access granted to %s for package %s", email, cli.GetPackageName())
	return output.Print(map[string]interface{}{
		"email":   email,
		"package": cli.GetPackageName(),
		"role":    role,
		"grant":   created.Name,
	})
}

func runRevoke(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("use --confirm to revoke access for %s", email)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would revoke access from %s for package %s", email, cli.GetPackageName())
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// The grant name format is: developers/{developer_id}/users/{email}/grants/{package_name}
	grantName := fmt.Sprintf("developers/-/users/%s/grants/%s", email, cli.GetPackageName())

	err = client.Grants().Delete(grantName).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Access revoked from %s for package %s", email, cli.GetPackageName())
	return nil
}
