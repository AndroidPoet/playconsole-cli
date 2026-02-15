package products

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var ProductsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage in-app products",
	Long: `Manage in-app products (managed products/one-time purchases).

In-app products are items that users can purchase within your app,
such as virtual goods, premium features, or consumable items.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all in-app products",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an in-app product",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an in-app product",
	RunE:  runCreate,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an in-app product",
	RunE:  runUpdate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an in-app product",
	RunE:  runDelete,
}

var (
	sku         string
	filePath    string
	priceUSD    string
	title       string
	description string
	maxResults  int64
	startIndex  int64
)

func init() {
	// List flags
	listCmd.Flags().Int64Var(&maxResults, "max-results", 100, "maximum results")
	listCmd.Flags().Int64Var(&startIndex, "start-index", 0, "starting index")

	// Get flags
	getCmd.Flags().StringVar(&sku, "sku", "", "product SKU")
	getCmd.MarkFlagRequired("sku")

	// Create flags
	createCmd.Flags().StringVar(&sku, "sku", "", "product SKU")
	createCmd.Flags().StringVar(&filePath, "file", "", "JSON file with product definition")
	createCmd.Flags().StringVar(&title, "title", "", "product title")
	createCmd.Flags().StringVar(&description, "description", "", "product description")
	createCmd.Flags().StringVar(&priceUSD, "price-usd", "", "price in USD (e.g., 0.99)")
	createCmd.MarkFlagRequired("sku")

	// Update flags
	updateCmd.Flags().StringVar(&sku, "sku", "", "product SKU")
	updateCmd.Flags().StringVar(&filePath, "file", "", "JSON file with product definition")
	updateCmd.Flags().StringVar(&title, "title", "", "product title")
	updateCmd.Flags().StringVar(&description, "description", "", "product description")
	updateCmd.Flags().StringVar(&priceUSD, "price-usd", "", "price in USD")
	updateCmd.MarkFlagRequired("sku")

	// Delete flags
	deleteCmd.Flags().StringVar(&sku, "sku", "", "product SKU")
	deleteCmd.Flags().Bool("confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("sku")

	ProductsCmd.AddCommand(listCmd)
	ProductsCmd.AddCommand(getCmd)
	ProductsCmd.AddCommand(createCmd)
	ProductsCmd.AddCommand(updateCmd)
	ProductsCmd.AddCommand(deleteCmd)
}

// ProductInfo represents product information
type ProductInfo struct {
	SKU              string `json:"sku"`
	Status           string `json:"status"`
	PurchaseType     string `json:"purchase_type"`
	DefaultPrice     string `json:"default_price,omitempty"`
	DefaultLanguage  string `json:"default_language,omitempty"`
	Title            string `json:"title,omitempty"`
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

	call := client.InAppProducts().List(client.GetPackageName()).Context(ctx)
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}
	if startIndex > 0 {
		call = call.StartIndex(startIndex)
	}

	products, err := call.Do()
	if err != nil {
		return err
	}

	result := make([]ProductInfo, 0, len(products.Inappproduct))
	for _, p := range products.Inappproduct {
		info := ProductInfo{
			SKU:             p.Sku,
			Status:          p.Status,
			PurchaseType:    p.PurchaseType,
			DefaultLanguage: p.DefaultLanguage,
		}

		if p.DefaultPrice != nil {
			info.DefaultPrice = fmt.Sprintf("%s %s", p.DefaultPrice.Currency, p.DefaultPrice.PriceMicros)
		}

		// Get title from listings
		if p.Listings != nil {
			for _, listing := range p.Listings {
				info.Title = listing.Title
				break
			}
		}

		result = append(result, info)
	}

	if len(result) == 0 {
		output.PrintInfo("No in-app products found")
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

	product, err := client.InAppProducts().Get(client.GetPackageName(), sku).Context(ctx).Do()
	if err != nil {
		return err
	}

	return output.Print(product)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	var product *androidpublisher.InAppProduct

	if filePath != "" {
		// Load from file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		product = &androidpublisher.InAppProduct{}
		if err := json.Unmarshal(data, product); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		// Build from flags
		if title == "" {
			return fmt.Errorf("--title is required when not using --file")
		}

		product = &androidpublisher.InAppProduct{
			Sku:             sku,
			Status:          "active",
			PurchaseType:    "managedUser",
			DefaultLanguage: "en-US",
			Listings: map[string]androidpublisher.InAppProductListing{
				"en-US": {
					Title:       title,
					Description: description,
				},
			},
		}

		if priceUSD != "" {
			// Convert to micros (1 USD = 1000000 micros)
			// This is simplified - real implementation would parse the price properly
			product.DefaultPrice = &androidpublisher.Price{
				Currency:    "USD",
				PriceMicros: priceUSD + "000000",
			}
		}
	}

	product.Sku = sku // Ensure SKU matches flag

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create product '%s'", sku)
		return output.Print(product)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	created, err := client.InAppProducts().Insert(client.GetPackageName(), product).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Product created: %s", sku)
	return output.Print(created)
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

	var product *androidpublisher.InAppProduct

	if filePath != "" {
		// Load from file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		product = &androidpublisher.InAppProduct{}
		if err := json.Unmarshal(data, product); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		// Get existing and update fields
		existing, err := client.InAppProducts().Get(client.GetPackageName(), sku).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to get existing product: %w", err)
		}
		product = existing

		if title != "" {
			if product.Listings == nil {
				product.Listings = make(map[string]androidpublisher.InAppProductListing)
			}
			listing := product.Listings[product.DefaultLanguage]
			listing.Title = title
			if description != "" {
				listing.Description = description
			}
			product.Listings[product.DefaultLanguage] = listing
		}

		if priceUSD != "" {
			product.DefaultPrice = &androidpublisher.Price{
				Currency:    "USD",
				PriceMicros: priceUSD + "000000",
			}
		}
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update product '%s'", sku)
		return output.Print(product)
	}

	updated, err := client.InAppProducts().Update(client.GetPackageName(), sku, product).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Product updated: %s", sku)
	return output.Print(updated)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("use --confirm to delete product '%s'", sku)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete product '%s'", sku)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	err = client.InAppProducts().Delete(client.GetPackageName(), sku).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Product deleted: %s", sku)
	return nil
}
