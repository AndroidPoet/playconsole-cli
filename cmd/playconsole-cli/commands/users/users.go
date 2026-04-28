package users

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
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
	developerID string
	email       string
	role        string
)

var rolePermissions = map[string][]string{
	"admin": {
		"CAN_MANAGE_PERMISSIONS",
	},
	"releaseManager": {
		"CAN_VIEW_NON_FINANCIAL_DATA",
		"CAN_VIEW_APP_QUALITY",
		"CAN_MANAGE_PUBLIC_APKS",
		"CAN_MANAGE_TRACK_APKS",
	},
	"appOwner": {
		"CAN_VIEW_NON_FINANCIAL_DATA",
		"CAN_VIEW_APP_QUALITY",
		"CAN_MANAGE_PUBLIC_APKS",
		"CAN_MANAGE_TRACK_APKS",
		"CAN_MANAGE_TRACK_USERS",
		"CAN_MANAGE_PUBLIC_LISTING",
		"CAN_REPLY_TO_REVIEWS",
		"CAN_MANAGE_APP_CONTENT",
		"CAN_MANAGE_DEEPLINKS",
	},
}

func init() {
	UsersCmd.PersistentFlags().StringVar(&developerID, "developer", "", "developer account ID")
	cli.MustMarkPersistentFlagRequired(UsersCmd, "developer")

	grantCmd.Flags().StringVar(&email, "email", "", "user email")
	grantCmd.Flags().StringVar(&role, "role", "releaseManager", "role: admin, releaseManager, appOwner")
	cli.MustMarkFlagRequired(grantCmd, "email")

	revokeCmd.Flags().StringVar(&email, "email", "", "user email")
	revokeCmd.Flags().Bool("confirm", false, "confirm revocation")
	cli.MustMarkFlagRequired(revokeCmd, "email")

	UsersCmd.AddCommand(listCmd)
	UsersCmd.AddCommand(grantCmd)
	UsersCmd.AddCommand(revokeCmd)
}

type GrantInfo struct {
	Name        string   `json:"name"`
	PackageName string   `json:"package_name,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type UserInfo struct {
	Email                       string      `json:"email"`
	Name                        string      `json:"name,omitempty"`
	AccessState                 string      `json:"access_state,omitempty"`
	DeveloperAccountPermissions []string    `json:"developer_account_permissions,omitempty"`
	Partial                     bool        `json:"partial,omitempty"`
	Grants                      []GrantInfo `json:"grants,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := api.NewClient("", 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	users, err := client.Users().List(developerParent()).Context(ctx).PageSize(-1).Do()
	if err != nil {
		return err
	}

	result := make([]UserInfo, 0, len(users.Users))
	for _, u := range users.Users {
		result = append(result, userInfoFromAPI(u))
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

	permissions, err := permissionsForRole(role)
	if err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo(
			"Dry run: would grant role '%s' to %s for package %s in developer %s",
			role, email, cli.GetPackageName(), developerID,
		)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	grant := &androidpublisher.Grant{
		PackageName:         cli.GetPackageName(),
		AppLevelPermissions: permissions,
	}

	parent := fmt.Sprintf("%s/users/%s", developerParent(), email)
	created, err := client.Grants().Create(parent, grant).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Access granted to %s for package %s", email, cli.GetPackageName())
	return output.Print(map[string]interface{}{
		"email":       email,
		"package":     cli.GetPackageName(),
		"role":        role,
		"permissions": permissions,
		"grant":       created.Name,
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
		output.PrintInfo(
			"Dry run: would revoke access from %s for package %s in developer %s",
			email, cli.GetPackageName(), developerID,
		)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	grantName := fmt.Sprintf("%s/users/%s/grants/%s", developerParent(), email, cli.GetPackageName())
	if err := client.Grants().Delete(grantName).Context(ctx).Do(); err != nil {
		return err
	}

	output.PrintSuccess("Access revoked from %s for package %s", email, cli.GetPackageName())
	return nil
}

func developerParent() string {
	return fmt.Sprintf("developers/%s", developerID)
}

func permissionsForRole(name string) ([]string, error) {
	permissions, ok := rolePermissions[name]
	if !ok {
		roles := make([]string, 0, len(rolePermissions))
		for roleName := range rolePermissions {
			roles = append(roles, roleName)
		}
		return nil, fmt.Errorf("invalid role %q. Valid roles: %s", name, strings.Join(roles, ", "))
	}

	return permissions, nil
}

func userInfoFromAPI(user *androidpublisher.User) UserInfo {
	info := UserInfo{
		Email:                       user.Email,
		Name:                        user.Name,
		AccessState:                 user.AccessState,
		DeveloperAccountPermissions: user.DeveloperAccountPermissions,
		Partial:                     user.Partial,
	}

	if len(user.Grants) > 0 {
		info.Grants = make([]GrantInfo, 0, len(user.Grants))
		for _, grant := range user.Grants {
			info.Grants = append(info.Grants, GrantInfo{
				Name:        grant.Name,
				PackageName: grant.PackageName,
				Permissions: grant.AppLevelPermissions,
			})
		}
	}

	return info
}
