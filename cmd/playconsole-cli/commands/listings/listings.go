package listings

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var ListingsCmd = &cobra.Command{
	Use:   "listings",
	Short: "Manage store listings",
	Long: `Manage localized store listings for your app.

Store listings include the app title, description, and other metadata
that appears on the Google Play Store page.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all localizations",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific localization",
	RunE:  runGet,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a store listing",
	RunE:  runUpdate,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync listings from a directory",
	Long: `Sync store listings from a directory structure.

Expected structure (fastlane-compatible):
  metadata/
    en-US/
      title.txt
      short_description.txt
      full_description.txt
    es-ES/
      title.txt
      ...`,
	RunE: runSync,
}

var (
	locale       string
	title        string
	shortDesc    string
	fullDesc     string
	fullDescFile string
	syncDir      string
)

func init() {
	// Get flags
	getCmd.Flags().StringVar(&locale, "locale", "", "locale code (e.g., en-US)")
	cli.MustMarkFlagRequired(getCmd, "locale")

	// Update flags
	updateCmd.Flags().StringVar(&locale, "locale", "", "locale code")
	updateCmd.Flags().StringVar(&title, "title", "", "app title")
	updateCmd.Flags().StringVar(&shortDesc, "short-description", "", "short description (80 chars)")
	updateCmd.Flags().StringVar(&fullDesc, "full-description", "", "full description")
	updateCmd.Flags().StringVar(&fullDescFile, "full-description-file", "", "file containing full description")
	cli.MustMarkFlagRequired(updateCmd, "locale")

	// Sync flags
	syncCmd.Flags().StringVar(&syncDir, "dir", "", "directory containing metadata")
	cli.MustMarkFlagRequired(syncCmd, "dir")

	ListingsCmd.AddCommand(listCmd)
	ListingsCmd.AddCommand(getCmd)
	ListingsCmd.AddCommand(updateCmd)
	ListingsCmd.AddCommand(syncCmd)
}

// ListingInfo represents a store listing
type ListingInfo struct {
	Locale           string `json:"locale"`
	Title            string `json:"title"`
	ShortDescription string `json:"short_description"`
	FullDescription  string `json:"full_description,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	listings, err := edit.Listings().List(client.GetPackageName(), edit.ID()).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	result := make([]ListingInfo, 0, len(listings.Listings))
	for _, l := range listings.Listings {
		result = append(result, ListingInfo{
			Locale:           l.Language,
			Title:            l.Title,
			ShortDescription: l.ShortDescription,
		})
	}

	if len(result) == 0 {
		output.PrintInfo("No listings found")
	}

	return output.Print(result)
}

// getOrNewListing fetches the existing listing for a locale. A 404 yields a
// fresh listing for that locale; any other error is returned as-is.
func getOrNewListing(client *api.Client, edit *api.Edit, loc string) (*androidpublisher.Listing, error) {
	existing, err := edit.Listings().Get(client.GetPackageName(), edit.ID(), loc).Context(edit.Context()).Do()
	if err == nil {
		return existing, nil
	}
	var gErr *googleapi.Error
	if errors.As(err, &gErr) && gErr.Code == http.StatusNotFound {
		return &androidpublisher.Listing{Language: loc}, nil
	}
	return nil, fmt.Errorf("failed to get listing for locale '%s': %w", loc, err)
}

func runGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	listing, err := edit.Listings().Get(client.GetPackageName(), edit.ID(), locale).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	return output.Print(ListingInfo{
		Locale:           listing.Language,
		Title:            listing.Title,
		ShortDescription: listing.ShortDescription,
		FullDescription:  listing.FullDescription,
	})
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if title == "" && shortDesc == "" && fullDesc == "" && fullDescFile == "" {
		return fmt.Errorf("nothing to update: provide at least one of --title, --short-description, --full-description, or --full-description-file")
	}

	// Read full description from file if specified
	desc := fullDesc
	if fullDescFile != "" {
		data, err := os.ReadFile(fullDescFile)
		if err != nil {
			return fmt.Errorf("failed to read description file: %w", err)
		}
		desc = strings.TrimRight(string(data), "\r\n")
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	ctx := edit.Context()

	// Get existing listing to preserve unchanged fields
	existing, err := getOrNewListing(client, edit, locale)
	if err != nil {
		return err
	}

	// Update fields if provided
	if title != "" {
		existing.Title = title
	}
	if shortDesc != "" {
		existing.ShortDescription = shortDesc
	}
	if desc != "" {
		existing.FullDescription = desc
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update listing for locale '%s'", locale)
		return output.Print(existing)
	}

	updated, err := edit.Listings().Update(client.GetPackageName(), edit.ID(), locale, existing).Context(ctx).Do()
	if err != nil {
		return err
	}

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Listing updated for locale '%s'", locale)
	return output.Print(ListingInfo{
		Locale:           updated.Language,
		Title:            updated.Title,
		ShortDescription: updated.ShortDescription,
		FullDescription:  updated.FullDescription,
	})
}

func runSync(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	absDir, err := filepath.Abs(syncDir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return fmt.Errorf("directory not found: %s", absDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", absDir)
	}

	// Read directory structure
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 2*time.Minute)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	ctx := edit.Context()
	updated := 0
	failures := make([]string, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		localeDir := filepath.Join(absDir, entry.Name())
		localeName := entry.Name()

		var newTitle, newShort, newFull *string

		// Read title
		if data, err := os.ReadFile(filepath.Join(localeDir, "title.txt")); err == nil {
			s := strings.TrimSpace(string(data))
			newTitle = &s
		}

		// Read short description
		if data, err := os.ReadFile(filepath.Join(localeDir, "short_description.txt")); err == nil {
			s := strings.TrimSpace(string(data))
			newShort = &s
		}

		// Read full description
		if data, err := os.ReadFile(filepath.Join(localeDir, "full_description.txt")); err == nil {
			s := strings.TrimRight(string(data), "\r\n")
			newFull = &s
		}

		if newTitle == nil && newShort == nil && newFull == nil {
			output.PrintWarning("Skipping %s: no title.txt, short_description.txt, or full_description.txt found", localeName)
			continue
		}

		// Merge onto the existing listing so absent files leave fields untouched
		listing, err := getOrNewListing(client, edit, localeName)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		if newTitle != nil {
			listing.Title = *newTitle
		}
		if newShort != nil {
			listing.ShortDescription = *newShort
		}
		if newFull != nil {
			listing.FullDescription = *newFull
		}

		if cli.IsDryRun() {
			output.PrintInfo("Dry run: would sync listing for locale '%s'", localeName)
			continue
		}

		_, err = edit.Listings().Update(client.GetPackageName(), edit.ID(), localeName, listing).Context(ctx).Do()
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", localeName, err))
			continue
		}

		output.PrintInfo("Updated: %s", localeName)
		updated++
	}

	if len(failures) > 0 {
		return fmt.Errorf("listing sync aborted; no changes committed. Failures: %s", strings.Join(failures, "; "))
	}

	if !cli.IsDryRun() && updated > 0 {
		if err := edit.Commit(); err != nil {
			return err
		}
	}

	output.PrintSuccess("Synced %d locale(s)", updated)
	return nil
}
