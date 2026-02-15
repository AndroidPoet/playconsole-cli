package edits

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var EditsCmd = &cobra.Command{
	Use:   "edits",
	Short: "Manage edit sessions (advanced)",
	Long: `Manage edit sessions for fine-grained control over app changes.

Most commands automatically manage edits for you. Use these commands
when you need manual control over the edit lifecycle.

Workflow:
  1. gpc edits create    - Start a new edit session
  2. Make changes (uploads, listing updates, etc.) with --no-auto-commit
  3. gpc edits validate  - Validate changes before committing
  4. gpc edits commit    - Commit changes to make them live`,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new edit session",
	RunE:  runCreate,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an edit session",
	RunE:  runGet,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate an edit session",
	RunE:  runValidate,
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit an edit session",
	RunE:  runCommit,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an edit session without committing",
	RunE:  runDelete,
}

var editID string

func init() {
	// Get flags
	getCmd.Flags().StringVar(&editID, "edit-id", "", "edit ID")
	getCmd.MarkFlagRequired("edit-id")

	// Validate flags
	validateCmd.Flags().StringVar(&editID, "edit-id", "", "edit ID")
	validateCmd.MarkFlagRequired("edit-id")

	// Commit flags
	commitCmd.Flags().StringVar(&editID, "edit-id", "", "edit ID")
	commitCmd.MarkFlagRequired("edit-id")

	// Delete flags
	deleteCmd.Flags().StringVar(&editID, "edit-id", "", "edit ID")
	deleteCmd.Flags().Bool("confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("edit-id")

	EditsCmd.AddCommand(createCmd)
	EditsCmd.AddCommand(getCmd)
	EditsCmd.AddCommand(validateCmd)
	EditsCmd.AddCommand(commitCmd)
	EditsCmd.AddCommand(deleteCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create new edit session")
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	// Don't close/delete - we want to keep this edit open

	output.PrintSuccess("Edit session created")
	return output.Print(map[string]interface{}{
		"edit_id":     edit.ID(),
		"package":     cli.GetPackageName(),
		"expires_in":  "1 hour",
		"next_steps": []string{
			"Make changes with --no-auto-commit flag",
			fmt.Sprintf("Validate: gpc edits validate --edit-id %s", edit.ID()),
			fmt.Sprintf("Commit: gpc edits commit --edit-id %s", edit.ID()),
		},
	})
}

func runGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.GetEdit(editID)
	if err != nil {
		return err
	}
	defer edit.Close()

	return output.Print(map[string]interface{}{
		"edit_id": edit.ID(),
		"package": cli.GetPackageName(),
		"status":  "active",
	})
}

func runValidate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would validate edit '%s'", editID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.GetEdit(editID)
	if err != nil {
		return err
	}
	defer edit.Close()

	if err := edit.Validate(); err != nil {
		return err
	}

	output.PrintSuccess("Edit validated successfully")
	return output.Print(map[string]interface{}{
		"edit_id": editID,
		"valid":   true,
	})
}

func runCommit(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would commit edit '%s'", editID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.GetEdit(editID)
	if err != nil {
		return err
	}
	defer edit.Close()

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Edit committed successfully")
	return output.Print(map[string]interface{}{
		"edit_id":   editID,
		"committed": true,
	})
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("use --confirm to delete edit '%s'", editID)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete edit '%s'", editID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.GetEdit(editID)
	if err != nil {
		return err
	}
	defer edit.Close()

	if err := edit.Delete(); err != nil {
		return err
	}

	output.PrintSuccess("Edit deleted: %s", editID)
	return nil
}
