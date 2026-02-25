package offers

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

// OffersCmd manages subscription offers
var OffersCmd = &cobra.Command{
	Use:   "offers",
	Short: "Manage subscription offers",
	Long: `Manage subscription offers for base plans.

Offers let you create introductory pricing, free trials, and promotional
offers for subscription base plans. Each offer has phases that define
the pricing and duration.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List offers for a base plan",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get offer details",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a subscription offer",
	RunE:  runCreate,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a subscription offer",
	RunE:  runUpdate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a subscription offer",
	RunE:  runDelete,
}

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate a subscription offer",
	RunE:  runActivate,
}

var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate a subscription offer",
	RunE:  runDeactivate,
}

var (
	productID  string
	basePlanID string
	offerID    string
	filePath   string
)

func init() {
	// List flags
	listCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	listCmd.Flags().StringVar(&basePlanID, "base-plan", "", "base plan ID")
	listCmd.MarkFlagRequired("product-id")
	listCmd.MarkFlagRequired("base-plan")

	// Get flags
	getCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	getCmd.Flags().StringVar(&basePlanID, "base-plan", "", "base plan ID")
	getCmd.Flags().StringVar(&offerID, "offer-id", "", "offer ID")
	getCmd.MarkFlagRequired("product-id")
	getCmd.MarkFlagRequired("base-plan")
	getCmd.MarkFlagRequired("offer-id")

	// Create flags
	createCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	createCmd.Flags().StringVar(&basePlanID, "base-plan", "", "base plan ID")
	createCmd.Flags().StringVar(&filePath, "file", "", "JSON file with offer definition")
	createCmd.MarkFlagRequired("product-id")
	createCmd.MarkFlagRequired("base-plan")
	createCmd.MarkFlagRequired("file")

	// Update flags
	updateCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	updateCmd.Flags().StringVar(&basePlanID, "base-plan", "", "base plan ID")
	updateCmd.Flags().StringVar(&offerID, "offer-id", "", "offer ID")
	updateCmd.Flags().StringVar(&filePath, "file", "", "JSON file with offer definition")
	updateCmd.MarkFlagRequired("product-id")
	updateCmd.MarkFlagRequired("base-plan")
	updateCmd.MarkFlagRequired("offer-id")
	updateCmd.MarkFlagRequired("file")

	// Delete flags
	deleteCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	deleteCmd.Flags().StringVar(&basePlanID, "base-plan", "", "base plan ID")
	deleteCmd.Flags().StringVar(&offerID, "offer-id", "", "offer ID")
	deleteCmd.Flags().Bool("confirm", false, "confirm destructive operation")
	deleteCmd.MarkFlagRequired("product-id")
	deleteCmd.MarkFlagRequired("base-plan")
	deleteCmd.MarkFlagRequired("offer-id")

	// Activate flags
	activateCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	activateCmd.Flags().StringVar(&basePlanID, "base-plan", "", "base plan ID")
	activateCmd.Flags().StringVar(&offerID, "offer-id", "", "offer ID")
	activateCmd.MarkFlagRequired("product-id")
	activateCmd.MarkFlagRequired("base-plan")
	activateCmd.MarkFlagRequired("offer-id")

	// Deactivate flags
	deactivateCmd.Flags().StringVar(&productID, "product-id", "", "subscription product ID")
	deactivateCmd.Flags().StringVar(&basePlanID, "base-plan", "", "base plan ID")
	deactivateCmd.Flags().StringVar(&offerID, "offer-id", "", "offer ID")
	deactivateCmd.MarkFlagRequired("product-id")
	deactivateCmd.MarkFlagRequired("base-plan")
	deactivateCmd.MarkFlagRequired("offer-id")

	OffersCmd.AddCommand(listCmd)
	OffersCmd.AddCommand(getCmd)
	OffersCmd.AddCommand(createCmd)
	OffersCmd.AddCommand(updateCmd)
	OffersCmd.AddCommand(deleteCmd)
	OffersCmd.AddCommand(activateCmd)
	OffersCmd.AddCommand(deactivateCmd)
}

// OfferInfo represents offer summary information
type OfferInfo struct {
	OfferID    string `json:"offer_id"`
	BasePlanID string `json:"base_plan_id"`
	ProductID  string `json:"product_id"`
	State      string `json:"state"`
	Phases     int    `json:"phases"`
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

	resp, err := client.Monetization().Subscriptions.BasePlans.Offers.List(
		client.GetPackageName(), productID, basePlanID,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to list offers: %w", err)
	}

	if len(resp.SubscriptionOffers) == 0 {
		output.PrintInfo("No offers found for base plan '%s'", basePlanID)
		return nil
	}

	result := make([]OfferInfo, 0, len(resp.SubscriptionOffers))
	for _, o := range resp.SubscriptionOffers {
		result = append(result, OfferInfo{
			OfferID:    o.OfferId,
			BasePlanID: o.BasePlanId,
			ProductID:  o.ProductId,
			State:      o.State,
			Phases:     len(o.Phases),
		})
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

	offer, err := client.Monetization().Subscriptions.BasePlans.Offers.Get(
		client.GetPackageName(), productID, basePlanID, offerID,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get offer: %w", err)
	}

	return output.Print(offer)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var offer androidpublisher.SubscriptionOffer
	if err := json.Unmarshal(data, &offer); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create offer for base plan '%s'", basePlanID)
		return output.Print(offer)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	created, err := client.Monetization().Subscriptions.BasePlans.Offers.Create(
		client.GetPackageName(), productID, basePlanID, &offer,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	output.PrintSuccess("Offer created: %s", created.OfferId)
	return output.Print(created)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var offer androidpublisher.SubscriptionOffer
	if err := json.Unmarshal(data, &offer); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update offer '%s'", offerID)
		return output.Print(offer)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	updated, err := client.Monetization().Subscriptions.BasePlans.Offers.Patch(
		client.GetPackageName(), productID, basePlanID, offerID, &offer,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to update offer: %w", err)
	}

	output.PrintSuccess("Offer updated: %s", offerID)
	return output.Print(updated)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete offer '%s'", offerID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	err = client.Monetization().Subscriptions.BasePlans.Offers.Delete(
		client.GetPackageName(), productID, basePlanID, offerID,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete offer: %w", err)
	}

	output.PrintSuccess("Offer deleted: %s", offerID)
	return nil
}

func runActivate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would activate offer '%s'", offerID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	activated, err := client.Monetization().Subscriptions.BasePlans.Offers.Activate(
		client.GetPackageName(), productID, basePlanID, offerID,
		&androidpublisher.ActivateSubscriptionOfferRequest{},
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to activate offer: %w", err)
	}

	output.PrintSuccess("Offer activated: %s", offerID)
	return output.Print(activated)
}

func runDeactivate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would deactivate offer '%s'", offerID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	deactivated, err := client.Monetization().Subscriptions.BasePlans.Offers.Deactivate(
		client.GetPackageName(), productID, basePlanID, offerID,
		&androidpublisher.DeactivateSubscriptionOfferRequest{},
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to deactivate offer: %w", err)
	}

	output.PrintSuccess("Offer deactivated: %s", offerID)
	return output.Print(deactivated)
}
