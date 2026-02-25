package orders

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// OrdersCmd manages orders and refunds
var OrdersCmd = &cobra.Command{
	Use:   "orders",
	Short: "Manage orders and refunds",
	Long: `View order details and issue refunds.

Orders represent completed purchases in your app, including
one-time purchases and subscription payments.`,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get order details",
	RunE:  runGet,
}

var refundCmd = &cobra.Command{
	Use:   "refund",
	Short: "Refund an order",
	RunE:  runRefund,
}

var batchGetCmd = &cobra.Command{
	Use:   "batch-get",
	Short: "Get multiple orders at once",
	RunE:  runBatchGet,
}

var (
	orderID  string
	orderIDs string
)

func init() {
	getCmd.Flags().StringVar(&orderID, "order-id", "", "order ID (e.g., GPA.1234-5678)")
	getCmd.MarkFlagRequired("order-id")

	refundCmd.Flags().StringVar(&orderID, "order-id", "", "order ID to refund")
	refundCmd.Flags().Bool("confirm", false, "confirm destructive operation")
	refundCmd.MarkFlagRequired("order-id")

	batchGetCmd.Flags().StringVar(&orderIDs, "order-ids", "", "comma-separated order IDs")
	batchGetCmd.MarkFlagRequired("order-ids")

	OrdersCmd.AddCommand(getCmd)
	OrdersCmd.AddCommand(refundCmd)
	OrdersCmd.AddCommand(batchGetCmd)
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

	order, err := client.Orders().Get(client.GetPackageName(), orderID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	return output.Print(order)
}

func runRefund(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would refund order '%s'", orderID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	err = client.Orders().Refund(client.GetPackageName(), orderID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to refund order: %w", err)
	}

	output.PrintSuccess("Order refunded: %s", orderID)
	return nil
}

func runBatchGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	ids := strings.Split(orderIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	resp, err := client.Orders().Batchget(client.GetPackageName()).OrderIds(ids...).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to batch get orders: %w", err)
	}

	return output.Print(resp)
}
