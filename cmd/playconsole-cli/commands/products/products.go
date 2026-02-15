package products

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

var ProductsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage in-app products (one-time purchases)",
	Long: `Manage in-app products (one-time purchases).

One-time products are items that users can purchase within your app,
such as virtual goods, premium features, or consumable items.

Uses the new Monetization API (monetization.onetimeproducts).`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all one-time products",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a one-time product",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a one-time product",
	RunE:  runCreate,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a one-time product",
	RunE:  runUpdate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a one-time product",
	RunE:  runDelete,
}

var (
	productID   string
	filePath    string
	title       string
	description string
	pageSize    int64
	pageToken   string
)

func init() {
	// List flags
	listCmd.Flags().Int64Var(&pageSize, "page-size", 100, "maximum results per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "page token for pagination")

	// Get flags
	getCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	getCmd.MarkFlagRequired("product-id")

	// Create flags
	createCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	createCmd.Flags().StringVar(&filePath, "file", "", "JSON file with product definition")
	createCmd.Flags().StringVar(&title, "title", "", "product title")
	createCmd.Flags().StringVar(&description, "description", "", "product description")
	createCmd.MarkFlagRequired("product-id")

	// Update flags
	updateCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	updateCmd.Flags().StringVar(&filePath, "file", "", "JSON file with product definition")
	updateCmd.Flags().StringVar(&title, "title", "", "product title")
	updateCmd.Flags().StringVar(&description, "description", "", "product description")
	updateCmd.MarkFlagRequired("product-id")

	// Delete flags
	deleteCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	deleteCmd.Flags().Bool("confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("product-id")

	ProductsCmd.AddCommand(listCmd)
	ProductsCmd.AddCommand(getCmd)
	ProductsCmd.AddCommand(createCmd)
	ProductsCmd.AddCommand(updateCmd)
	ProductsCmd.AddCommand(deleteCmd)
}

// ProductInfo represents product information
type ProductInfo struct {
	ProductID   string `json:"product_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
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

	call := client.Monetization().Onetimeproducts.List(client.GetPackageName()).Context(ctx)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	products, err := call.Do()
	if err != nil {
		return err
	}

	result := make([]ProductInfo, 0, len(products.OneTimeProducts))
	for _, p := range products.OneTimeProducts {
		info := ProductInfo{
			ProductID: p.ProductId,
		}

		// Get title and description from listings
		if len(p.Listings) > 0 {
			info.Title = p.Listings[0].Title
			info.Description = p.Listings[0].Description
		}

		result = append(result, info)
	}

	if len(result) == 0 {
		output.PrintInfo("No one-time products found")
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

	product, err := client.Monetization().Onetimeproducts.Get(client.GetPackageName(), productID).Context(ctx).Do()
	if err != nil {
		return err
	}

	return output.Print(product)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	var product *androidpublisher.OneTimeProduct

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		product = &androidpublisher.OneTimeProduct{}
		if err := json.Unmarshal(data, product); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		// Create minimal product from flags
		product = &androidpublisher.OneTimeProduct{
			PackageName: cli.GetPackageName(),
			ProductId:   productID,
			Listings: []*androidpublisher.OneTimeProductListing{
				{
					LanguageCode: "en-US",
					Title:        title,
					Description:  description,
				},
			},
		}
	}

	// Ensure package name and product ID are set
	product.PackageName = cli.GetPackageName()
	product.ProductId = productID

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create product")
		return output.Print(product)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Use Patch with allowMissing=true to create
	result, err := client.Monetization().Onetimeproducts.Patch(client.GetPackageName(), productID, product).
		AllowMissing(true).
		Context(ctx).
		Do()
	if err != nil {
		return err
	}

	return output.Print(result)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	var product *androidpublisher.OneTimeProduct

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		product = &androidpublisher.OneTimeProduct{}
		if err := json.Unmarshal(data, product); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		// Get existing product first
		existing, err := client.Monetization().Onetimeproducts.Get(client.GetPackageName(), productID).Context(ctx).Do()
		if err != nil {
			return err
		}
		product = existing

		// Update fields if provided
		if title != "" || description != "" {
			if len(product.Listings) == 0 {
				product.Listings = []*androidpublisher.OneTimeProductListing{{LanguageCode: "en-US"}}
			}
			if title != "" {
				product.Listings[0].Title = title
			}
			if description != "" {
				product.Listings[0].Description = description
			}
		}
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update product")
		return output.Print(product)
	}

	result, err := client.Monetization().Onetimeproducts.Patch(client.GetPackageName(), productID, product).
		Context(ctx).
		Do()
	if err != nil {
		return err
	}

	return output.Print(result)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("deletion requires --confirm flag")
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete product %s", productID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	err = client.Monetization().Onetimeproducts.Delete(client.GetPackageName(), productID).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintInfo("Product '%s' deleted", productID)
	return nil
}
