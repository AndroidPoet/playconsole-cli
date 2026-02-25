package recovery

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// RecoveryCmd manages app recovery actions
var RecoveryCmd = &cobra.Command{
	Use:   "recovery",
	Short: "Manage app recovery actions",
	Long: `Create and manage app recovery actions for production incidents.

Recovery actions let you remotely trigger remediation for affected
users, such as clearing app data or prompting an update, without
requiring a new app release.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recovery actions",
	RunE:  runList,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a draft recovery action",
	RunE:  runCreate,
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a recovery action",
	RunE:  runDeploy,
}

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a recovery action",
	RunE:  runCancel,
}

var addTargetingCmd = &cobra.Command{
	Use:   "add-targeting",
	Short: "Add targeting to a recovery action",
	RunE:  runAddTargeting,
}

var (
	recoveryID int64
	filePath   string
)

func init() {
	createCmd.Flags().StringVar(&filePath, "file", "", "JSON file with recovery action definition")
	createCmd.MarkFlagRequired("file")

	deployCmd.Flags().Int64Var(&recoveryID, "recovery-id", 0, "recovery action ID")
	deployCmd.Flags().Bool("confirm", false, "confirm destructive operation")
	deployCmd.MarkFlagRequired("recovery-id")

	cancelCmd.Flags().Int64Var(&recoveryID, "recovery-id", 0, "recovery action ID")
	cancelCmd.Flags().Bool("confirm", false, "confirm destructive operation")
	cancelCmd.MarkFlagRequired("recovery-id")

	addTargetingCmd.Flags().Int64Var(&recoveryID, "recovery-id", 0, "recovery action ID")
	addTargetingCmd.Flags().StringVar(&filePath, "file", "", "JSON file with targeting definition")
	addTargetingCmd.MarkFlagRequired("recovery-id")
	addTargetingCmd.MarkFlagRequired("file")

	RecoveryCmd.AddCommand(listCmd)
	RecoveryCmd.AddCommand(createCmd)
	RecoveryCmd.AddCommand(deployCmd)
	RecoveryCmd.AddCommand(cancelCmd)
	RecoveryCmd.AddCommand(addTargetingCmd)
}

// RecoveryInfo represents recovery action summary
type RecoveryInfo struct {
	RecoveryID int64  `json:"recovery_id"`
	Status     string `json:"status"`
	AppVersion int64  `json:"app_version,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	resp, err := client.AppRecovery().List(client.GetPackageName()).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to list recovery actions: %w", err)
	}

	if len(resp.RecoveryActions) == 0 {
		output.PrintInfo("No recovery actions found")
		return nil
	}

	result := make([]RecoveryInfo, 0, len(resp.RecoveryActions))
	for _, r := range resp.RecoveryActions {
		result = append(result, RecoveryInfo{
			RecoveryID: r.AppRecoveryId,
			Status:     r.Status,
		})
	}

	return output.Print(result)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var req androidpublisher.CreateDraftAppRecoveryRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create recovery action")
		return output.Print(req)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	created, err := client.AppRecovery().Create(client.GetPackageName(), &req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create recovery action: %w", err)
	}

	output.PrintSuccess("Recovery action created")
	return output.Print(created)
}

func runDeploy(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would deploy recovery action %d", recoveryID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	deployed, err := client.AppRecovery().Deploy(
		client.GetPackageName(), recoveryID,
		&androidpublisher.DeployAppRecoveryRequest{},
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to deploy recovery action: %w", err)
	}

	output.PrintSuccess("Recovery action %d deployed", recoveryID)
	return output.Print(deployed)
}

func runCancel(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would cancel recovery action %d", recoveryID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	cancelled, err := client.AppRecovery().Cancel(
		client.GetPackageName(), recoveryID,
		&androidpublisher.CancelAppRecoveryRequest{},
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to cancel recovery action: %w", err)
	}

	output.PrintSuccess("Recovery action %d cancelled", recoveryID)
	return output.Print(cancelled)
}

func runAddTargeting(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var req androidpublisher.AddTargetingRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would add targeting to recovery action %d", recoveryID)
		return output.Print(req)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	resp, err := client.AppRecovery().AddTargeting(
		client.GetPackageName(), recoveryID, &req,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to add targeting: %w", err)
	}

	output.PrintSuccess("Targeting added to recovery action %d", recoveryID)
	return output.Print(resp)
}
