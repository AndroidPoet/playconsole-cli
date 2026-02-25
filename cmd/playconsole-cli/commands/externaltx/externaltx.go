package externaltx

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

// ExternalTxCmd manages external transactions
var ExternalTxCmd = &cobra.Command{
	Use:     "external-transactions",
	Aliases: []string{"ext-tx"},
	Short:   "Manage external transactions",
	Long: `Manage transactions processed outside of Google Play Billing.

External transactions are used for alternative billing compliance,
allowing you to report purchases made through third-party payment
systems.`,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an external transaction",
	RunE:  runCreate,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get external transaction details",
	RunE:  runGet,
}

var refundCmd = &cobra.Command{
	Use:   "refund",
	Short: "Refund an external transaction",
	RunE:  runRefund,
}

var (
	filePath string
	txName   string
)

func init() {
	createCmd.Flags().StringVar(&filePath, "file", "", "JSON file with transaction definition")
	createCmd.MarkFlagRequired("file")

	getCmd.Flags().StringVar(&txName, "name", "", "transaction resource name")
	getCmd.MarkFlagRequired("name")

	refundCmd.Flags().StringVar(&txName, "name", "", "transaction resource name")
	refundCmd.Flags().StringVar(&filePath, "file", "", "JSON file with refund request (optional)")
	refundCmd.Flags().Bool("confirm", false, "confirm destructive operation")
	refundCmd.MarkFlagRequired("name")

	ExternalTxCmd.AddCommand(createCmd)
	ExternalTxCmd.AddCommand(getCmd)
	ExternalTxCmd.AddCommand(refundCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var tx androidpublisher.ExternalTransaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create external transaction")
		return output.Print(tx)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	parent := fmt.Sprintf("applications/%s", client.GetPackageName())
	created, err := client.ExternalTransactions().Createexternaltransaction(parent, &tx).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create external transaction: %w", err)
	}

	output.PrintSuccess("External transaction created")
	return output.Print(created)
}

func runGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	tx, err := client.ExternalTransactions().Getexternaltransaction(txName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get external transaction: %w", err)
	}

	return output.Print(tx)
}

func runRefund(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}

	var req androidpublisher.RefundExternalTransactionRequest
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would refund external transaction '%s'", txName)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	refunded, err := client.ExternalTransactions().Refundexternaltransaction(txName, &req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to refund external transaction: %w", err)
	}

	output.PrintSuccess("External transaction refunded")
	return output.Print(refunded)
}
