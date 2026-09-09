package products

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
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
	updateMask  string
	pageSize    int64
)

// fullUpdateMask covers every writable top-level field of a one-time product.
const fullUpdateMask = "listings,purchaseOptions,offerTags,taxAndComplianceSettings,restrictedPaymentCountries"

func init() {
	// List flags
	listCmd.Flags().Int64Var(&pageSize, "page-size", 100, "maximum results per page")

	// Get flags
	getCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	cli.MustMarkFlagRequired(getCmd, "product-id")

	// Create flags
	createCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	createCmd.Flags().StringVar(&filePath, "file", "", "JSON file with product definition (required; must include purchaseOptions)")
	createCmd.Flags().StringVar(&title, "title", "", "override the listing title from the file")
	createCmd.Flags().StringVar(&description, "description", "", "override the listing description from the file")
	createCmd.Flags().StringVar(&updateMask, "update-mask", "", "comma-separated field mask (default: derived from the file's top-level keys)")
	cli.MustMarkFlagRequired(createCmd, "product-id")

	// Update flags
	updateCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	updateCmd.Flags().StringVar(&filePath, "file", "", "JSON file with product definition")
	updateCmd.Flags().StringVar(&title, "title", "", "product title")
	updateCmd.Flags().StringVar(&description, "description", "", "product description")
	updateCmd.Flags().StringVar(&updateMask, "update-mask", "", "comma-separated field mask (default: derived from the file's top-level keys)")
	cli.MustMarkFlagRequired(updateCmd, "product-id")

	// Delete flags
	deleteCmd.Flags().StringVar(&productID, "product-id", "", "product ID")
	deleteCmd.Flags().Bool("confirm", false, "confirm deletion")
	cli.MustMarkFlagRequired(deleteCmd, "product-id")

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

	call := client.Monetization().Onetimeproducts.List(client.GetPackageName())
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}

	result := make([]ProductInfo, 0)
	err = call.Pages(ctx, func(page *androidpublisher.ListOneTimeProductsResponse) error {
		for _, p := range page.OneTimeProducts {
			info := ProductInfo{
				ProductID: p.ProductId,
			}

			if l := preferredListing(p.Listings); l != nil {
				info.Title = l.Title
				info.Description = l.Description
			}

			result = append(result, info)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(result) == 0 {
		output.PrintInfo("No one-time products found")
	}

	return output.Print(result)
}

// preferredListing returns the en-US listing when present, otherwise the first one.
func preferredListing(listings []*androidpublisher.OneTimeProductListing) *androidpublisher.OneTimeProductListing {
	for _, l := range listings {
		if l != nil && l.LanguageCode == "en-US" {
			return l
		}
	}
	if len(listings) > 0 {
		return listings[0]
	}
	return nil
}

// readProductFile parses a product definition and derives an update mask from
// the JSON's top-level keys (identifiers are excluded since they are not
// writable through the mask).
func readProductFile(path string) (*androidpublisher.OneTimeProduct, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	product := &androidpublisher.OneTimeProduct{}
	if err := json.Unmarshal(data, product); err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	keys := make([]string, 0, len(raw))
	for k := range raw {
		if k == "packageName" || k == "productId" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return product, strings.Join(keys, ","), nil
}

// applyListingOverrides sets title/description on the preferred listing,
// creating an en-US listing when the product has none. It reports whether
// anything was changed.
func applyListingOverrides(product *androidpublisher.OneTimeProduct) bool {
	if title == "" && description == "" {
		return false
	}
	listing := preferredListing(product.Listings)
	if listing == nil {
		listing = &androidpublisher.OneTimeProductListing{LanguageCode: "en-US"}
		product.Listings = append(product.Listings, listing)
	}
	if title != "" {
		listing.Title = title
	}
	if description != "" {
		listing.Description = description
	}
	return true
}

// withMaskField appends field to a comma-separated mask if it is not present.
func withMaskField(mask, field string) string {
	if mask == "" {
		return field
	}
	for _, f := range strings.Split(mask, ",") {
		if strings.TrimSpace(f) == field {
			return mask
		}
	}
	return mask + "," + field
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

	if filePath == "" {
		return fmt.Errorf("--file is required: the API needs purchaseOptions, which cannot be expressed with flags alone")
	}

	product, mask, err := readProductFile(filePath)
	if err != nil {
		return err
	}

	if applyListingOverrides(product) {
		mask = withMaskField(mask, "listings")
	}
	if updateMask != "" {
		mask = updateMask
	}
	if mask == "" {
		mask = fullUpdateMask
	}

	// Ensure package name and product ID are set
	product.PackageName = cli.GetPackageName()
	product.ProductId = productID

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create product (update mask: %s)", mask)
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
		UpdateMask(mask).
		RegionsVersionVersion("2022/02").
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
	var mask string

	if filePath != "" {
		product, mask, err = readProductFile(filePath)
		if err != nil {
			return err
		}
		if applyListingOverrides(product) {
			mask = withMaskField(mask, "listings")
		}
		product.PackageName = cli.GetPackageName()
		product.ProductId = productID
	} else {
		if title == "" && description == "" {
			return fmt.Errorf("nothing to update: provide --file, --title, or --description")
		}

		// Get existing product first
		existing, err := client.Monetization().Onetimeproducts.Get(client.GetPackageName(), productID).Context(ctx).Do()
		if err != nil {
			return err
		}
		product = existing
		applyListingOverrides(product)
		mask = "listings"
	}

	if updateMask != "" {
		mask = updateMask
	}
	if mask == "" {
		return fmt.Errorf("update mask is empty: the file has no updatable top-level fields; use --update-mask")
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update product (update mask: %s)", mask)
		return output.Print(product)
	}

	result, err := client.Monetization().Onetimeproducts.Patch(client.GetPackageName(), productID, product).
		UpdateMask(mask).
		RegionsVersionVersion("2022/02").
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

	output.PrintSuccess("Product '%s' deleted", productID)
	return output.Print(map[string]interface{}{
		"product_id": productID,
		"deleted":    true,
	})
}
