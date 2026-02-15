package purchases

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var PurchasesCmd = &cobra.Command{
	Use:   "purchases",
	Short: "Manage purchases and verify tokens",
	Long: `Verify and manage in-app purchases and subscriptions.

This allows you to verify purchase tokens, check subscription status,
acknowledge purchases, and handle refunds.`,
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a purchase token",
	RunE:  runVerify,
}

var subscriptionStatusCmd = &cobra.Command{
	Use:   "subscription-status",
	Short: "Get subscription status",
	RunE:  runSubscriptionStatus,
}

var acknowledgeCmd = &cobra.Command{
	Use:   "acknowledge",
	Short: "Acknowledge a purchase",
	RunE:  runAcknowledge,
}

var voidedCmd = &cobra.Command{
	Use:   "voided",
	Short: "List voided purchases",
}

var voidedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List voided purchases",
	RunE:  runVoidedList,
}

var (
	purchaseToken string
	productID     string
	startTime     string
	endTime       string
	maxResults    int64
)

func init() {
	// Verify flags
	verifyCmd.Flags().StringVar(&purchaseToken, "token", "", "purchase token")
	verifyCmd.Flags().StringVar(&productID, "product-id", "", "product ID (for products)")
	verifyCmd.MarkFlagRequired("token")

	// Subscription status flags
	subscriptionStatusCmd.Flags().StringVar(&purchaseToken, "token", "", "subscription token")
	subscriptionStatusCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	subscriptionStatusCmd.MarkFlagRequired("token")
	subscriptionStatusCmd.MarkFlagRequired("product-id")

	// Acknowledge flags
	acknowledgeCmd.Flags().StringVar(&purchaseToken, "token", "", "purchase token")
	acknowledgeCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	acknowledgeCmd.MarkFlagRequired("token")
	acknowledgeCmd.MarkFlagRequired("product-id")

	// Voided list flags
	voidedListCmd.Flags().StringVar(&startTime, "start-time", "", "start time (RFC3339)")
	voidedListCmd.Flags().StringVar(&endTime, "end-time", "", "end time (RFC3339)")
	voidedListCmd.Flags().Int64Var(&maxResults, "max-results", 100, "maximum results")

	voidedCmd.AddCommand(voidedListCmd)

	PurchasesCmd.AddCommand(verifyCmd)
	PurchasesCmd.AddCommand(subscriptionStatusCmd)
	PurchasesCmd.AddCommand(acknowledgeCmd)
	PurchasesCmd.AddCommand(voidedCmd)
}

func runVerify(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if productID == "" {
		return fmt.Errorf("--product-id is required to verify purchases")
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Verify as a product purchase
	purchase, err := client.Purchases().Products.Get(client.GetPackageName(), productID, purchaseToken).Context(ctx).Do()
	if err != nil {
		return err
	}

	return output.Print(map[string]interface{}{
		"type":              "product",
		"product_id":        productID,
		"purchase_state":    purchase.PurchaseState,
		"consumption_state": purchase.ConsumptionState,
		"acknowledged":      purchase.AcknowledgementState == 1,
		"purchase_time_ms":  purchase.PurchaseTimeMillis,
		"order_id":          purchase.OrderId,
	})
}

func runSubscriptionStatus(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Use the subscriptions v2 API
	sub, err := client.Purchases().Subscriptionsv2.Get(client.GetPackageName(), purchaseToken).Context(ctx).Do()
	if err != nil {
		return err
	}

	result := map[string]interface{}{
		"subscription_state": sub.SubscriptionState,
		"latest_order_id":    sub.LatestOrderId,
	}

	if sub.StartTime != "" {
		result["start_time"] = sub.StartTime
	}

	if len(sub.LineItems) > 0 {
		items := make([]map[string]interface{}, 0, len(sub.LineItems))
		for _, item := range sub.LineItems {
			items = append(items, map[string]interface{}{
				"product_id":    item.ProductId,
				"expiry_time":   item.ExpiryTime,
				"auto_renewing": item.AutoRenewingPlan != nil,
			})
		}
		result["line_items"] = items
	}

	return output.Print(result)
}

func runAcknowledge(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would acknowledge purchase for product '%s'", productID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	err = client.Purchases().Products.Acknowledge(client.GetPackageName(), productID, purchaseToken, nil).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Purchase acknowledged: %s", productID)
	return nil
}

func runVoidedList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	call := client.Purchases().Voidedpurchases.List(client.GetPackageName()).Context(ctx)

	if startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			return err
		}
		call = call.StartTime(t.UnixMilli())
	}

	if endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			return err
		}
		call = call.EndTime(t.UnixMilli())
	}

	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}

	voided, err := call.Do()
	if err != nil {
		return err
	}

	if len(voided.VoidedPurchases) == 0 {
		output.PrintInfo("No voided purchases found")
		return nil
	}

	type VoidedInfo struct {
		OrderID      string `json:"order_id"`
		VoidedTime   string `json:"voided_time"`
		VoidedSource int64  `json:"voided_source"`
		VoidedReason int64  `json:"voided_reason"`
	}

	result := make([]VoidedInfo, 0, len(voided.VoidedPurchases))
	for _, v := range voided.VoidedPurchases {
		result = append(result, VoidedInfo{
			OrderID:      v.OrderId,
			VoidedTime:   time.UnixMilli(v.VoidedTimeMillis).Format(time.RFC3339),
			VoidedSource: v.VoidedSource,
			VoidedReason: v.VoidedReason,
		})
	}

	return output.Print(result)
}
