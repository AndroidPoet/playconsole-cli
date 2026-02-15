package subscriptions

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

var SubscriptionsCmd = &cobra.Command{
	Use:   "subscriptions",
	Short: "Manage subscriptions",
	Long: `Manage subscription products and base plans.

Subscriptions provide recurring billing for your app's premium features
or content. Each subscription can have multiple base plans with different
billing periods and pricing.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all subscriptions",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get subscription details",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a subscription",
	RunE:  runCreate,
}

var basePlansCmd = &cobra.Command{
	Use:   "base-plans",
	Short: "Manage base plans",
}

var basePlansListCmd = &cobra.Command{
	Use:   "list",
	Short: "List base plans for a subscription",
	RunE:  runBasePlansList,
}

var basePlansCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a base plan",
	RunE:  runBasePlansCreate,
}

var pricingCmd = &cobra.Command{
	Use:   "pricing",
	Short: "Manage subscription pricing",
}

var pricingGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get pricing for a base plan",
	RunE:  runPricingGet,
}

var (
	productID  string
	basePlanID string
	filePath   string
	maxResults int64
)

func init() {
	// List flags
	listCmd.Flags().Int64Var(&maxResults, "max-results", 100, "maximum results")

	// Get flags
	getCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	getCmd.MarkFlagRequired("product-id")

	// Create flags
	createCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	createCmd.Flags().StringVar(&filePath, "file", "", "JSON file with subscription definition")
	createCmd.MarkFlagRequired("product-id")
	createCmd.MarkFlagRequired("file")

	// Base plans list flags
	basePlansListCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	basePlansListCmd.MarkFlagRequired("product-id")

	// Base plans create flags
	basePlansCreateCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	basePlansCreateCmd.Flags().StringVar(&filePath, "file", "", "JSON file with base plan definition")
	basePlansCreateCmd.MarkFlagRequired("product-id")
	basePlansCreateCmd.MarkFlagRequired("file")

	// Pricing get flags
	pricingGetCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	pricingGetCmd.Flags().StringVar(&basePlanID, "base-plan", "", "base plan ID")
	pricingGetCmd.MarkFlagRequired("product-id")
	pricingGetCmd.MarkFlagRequired("base-plan")

	// Build command tree
	basePlansCmd.AddCommand(basePlansListCmd)
	basePlansCmd.AddCommand(basePlansCreateCmd)

	pricingCmd.AddCommand(pricingGetCmd)

	SubscriptionsCmd.AddCommand(listCmd)
	SubscriptionsCmd.AddCommand(getCmd)
	SubscriptionsCmd.AddCommand(createCmd)
	SubscriptionsCmd.AddCommand(basePlansCmd)
	SubscriptionsCmd.AddCommand(pricingCmd)
}

// SubscriptionInfo represents subscription information
type SubscriptionInfo struct {
	ProductID   string `json:"product_id"`
	BasePlans   int    `json:"base_plans"`
	PackageName string `json:"package_name,omitempty"`
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

	subs, err := client.Monetization().Subscriptions.List(client.GetPackageName()).Context(ctx).Do()
	if err != nil {
		return err
	}

	result := make([]SubscriptionInfo, 0, len(subs.Subscriptions))
	for _, s := range subs.Subscriptions {
		result = append(result, SubscriptionInfo{
			ProductID:   s.ProductId,
			BasePlans:   len(s.BasePlans),
			PackageName: s.PackageName,
		})
	}

	if len(result) == 0 {
		output.PrintInfo("No subscriptions found")
		return nil
	}

	return output.Print(result)
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

	sub, err := client.Monetization().Subscriptions.Get(client.GetPackageName(), productID).Context(ctx).Do()
	if err != nil {
		return err
	}

	return output.Print(sub)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var sub androidpublisher.Subscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	sub.ProductId = productID
	sub.PackageName = cli.GetPackageName()

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create subscription '%s'", productID)
		return output.Print(sub)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	created, err := client.Monetization().Subscriptions.Create(client.GetPackageName(), &sub).ProductId(productID).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Subscription created: %s", productID)
	return output.Print(created)
}

func runBasePlansList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	sub, err := client.Monetization().Subscriptions.Get(client.GetPackageName(), productID).Context(ctx).Do()
	if err != nil {
		return err
	}

	if len(sub.BasePlans) == 0 {
		output.PrintInfo("No base plans found for subscription '%s'", productID)
		return nil
	}

	type BasePlanInfo struct {
		BasePlanID string `json:"base_plan_id"`
		State      string `json:"state"`
		Offers     int    `json:"offers"`
	}

	result := make([]BasePlanInfo, 0, len(sub.BasePlans))
	for _, bp := range sub.BasePlans {
		result = append(result, BasePlanInfo{
			BasePlanID: bp.BasePlanId,
			State:      bp.State,
			Offers:     len(bp.OfferTags),
		})
	}

	return output.Print(result)
}

func runBasePlansCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var basePlan androidpublisher.BasePlan
	if err := json.Unmarshal(data, &basePlan); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create base plan for subscription '%s'", productID)
		return output.Print(basePlan)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Get existing subscription
	sub, err := client.Monetization().Subscriptions.Get(client.GetPackageName(), productID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Add new base plan
	sub.BasePlans = append(sub.BasePlans, &basePlan)

	// Update subscription
	updated, err := client.Monetization().Subscriptions.Patch(client.GetPackageName(), productID, sub).UpdateMask("basePlans").Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Base plan created for subscription '%s'", productID)
	return output.Print(updated)
}

func runPricingGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	sub, err := client.Monetization().Subscriptions.Get(client.GetPackageName(), productID).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Find the base plan
	for _, bp := range sub.BasePlans {
		if bp.BasePlanId == basePlanID {
			return output.Print(map[string]interface{}{
				"base_plan_id": bp.BasePlanId,
				"state":        bp.State,
				"pricing":      bp.RegionalConfigs,
			})
		}
	}

	return fmt.Errorf("base plan '%s' not found in subscription '%s'", basePlanID, productID)
}
