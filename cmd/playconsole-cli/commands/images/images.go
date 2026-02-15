package images

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var ImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Manage screenshots and graphics",
	Long: `Manage screenshots, icons, and promotional graphics for your app.

Image types:
  - phoneScreenshots: Phone screenshots (up to 8)
  - sevenInchScreenshots: 7-inch tablet screenshots
  - tenInchScreenshots: 10-inch tablet screenshots
  - tvScreenshots: TV screenshots
  - wearScreenshots: Wear OS screenshots
  - featureGraphic: Feature graphic (1024x500)
  - icon: App icon (512x512)
  - tvBanner: TV banner (1280x720)
  - promoGraphic: Promo graphic (180x120)`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List images for a locale and type",
	RunE:  runList,
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an image",
	RunE:  runUpload,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an image",
	RunE:  runDelete,
}

var deleteAllCmd = &cobra.Command{
	Use:   "delete-all",
	Short: "Delete all images of a type",
	RunE:  runDeleteAll,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync images from a directory",
	Long: `Sync images from a directory structure.

Expected structure:
  screenshots/
    en-US/
      phoneScreenshots/
        1.png
        2.png
      featureGraphic/
        feature.png`,
	RunE: runSync,
}

var (
	locale    string
	imageType string
	imageID   string
	filePath  string
	syncDir   string
)

// Valid image types
var validImageTypes = []string{
	"phoneScreenshots",
	"sevenInchScreenshots",
	"tenInchScreenshots",
	"tvScreenshots",
	"wearScreenshots",
	"featureGraphic",
	"icon",
	"tvBanner",
	"promoGraphic",
}

func init() {
	// List flags
	listCmd.Flags().StringVar(&locale, "locale", "", "locale code (e.g., en-US)")
	listCmd.Flags().StringVar(&imageType, "type", "", "image type")
	listCmd.MarkFlagRequired("locale")
	listCmd.MarkFlagRequired("type")

	// Upload flags
	uploadCmd.Flags().StringVar(&locale, "locale", "", "locale code")
	uploadCmd.Flags().StringVar(&imageType, "type", "", "image type")
	uploadCmd.Flags().StringVar(&filePath, "file", "", "path to image file")
	uploadCmd.MarkFlagRequired("locale")
	uploadCmd.MarkFlagRequired("type")
	uploadCmd.MarkFlagRequired("file")

	// Delete flags
	deleteCmd.Flags().StringVar(&locale, "locale", "", "locale code")
	deleteCmd.Flags().StringVar(&imageType, "type", "", "image type")
	deleteCmd.Flags().StringVar(&imageID, "id", "", "image ID to delete")
	deleteCmd.Flags().Bool("confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("locale")
	deleteCmd.MarkFlagRequired("type")
	deleteCmd.MarkFlagRequired("id")

	// Delete all flags
	deleteAllCmd.Flags().StringVar(&locale, "locale", "", "locale code")
	deleteAllCmd.Flags().StringVar(&imageType, "type", "", "image type")
	deleteAllCmd.Flags().Bool("confirm", false, "confirm deletion")
	deleteAllCmd.MarkFlagRequired("locale")
	deleteAllCmd.MarkFlagRequired("type")

	// Sync flags
	syncCmd.Flags().StringVar(&syncDir, "dir", "", "directory containing images")
	syncCmd.MarkFlagRequired("dir")

	ImagesCmd.AddCommand(listCmd)
	ImagesCmd.AddCommand(uploadCmd)
	ImagesCmd.AddCommand(deleteCmd)
	ImagesCmd.AddCommand(deleteAllCmd)
	ImagesCmd.AddCommand(syncCmd)
}

// ImageInfo represents image information
type ImageInfo struct {
	ID  string `json:"id"`
	URL string `json:"url,omitempty"`
	SHA1 string `json:"sha1,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

func validateImageType(t string) error {
	for _, valid := range validImageTypes {
		if t == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid image type '%s'. Valid types: %s", t, strings.Join(validImageTypes, ", "))
}

func runList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := validateImageType(imageType); err != nil {
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
	defer edit.Delete()

	images, err := edit.Images().List(client.GetPackageName(), edit.ID(), locale, imageType).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	result := make([]ImageInfo, 0, len(images.Images))
	for _, img := range images.Images {
		result = append(result, ImageInfo{
			ID:  img.Id,
			URL: img.Url,
			SHA1: img.Sha1,
			SHA256: img.Sha256,
		})
	}

	if len(result) == 0 {
		output.PrintInfo("No images found for %s/%s", locale, imageType)
		return nil
	}

	return output.Print(result)
}

func runUpload(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := validateImageType(imageType); err != nil {
		return err
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("file not found: %s", absPath)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would upload %s to %s/%s", filepath.Base(absPath), locale, imageType)
		return nil
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

	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	output.PrintInfo("Uploading: %s (%d bytes)", filepath.Base(absPath), info.Size())

	image, err := edit.Images().Upload(client.GetPackageName(), edit.ID(), locale, imageType).Media(file).Context(edit.Context()).Do()
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Image uploaded: %s", image.Image.Id)
	return output.Print(ImageInfo{
		ID:  image.Image.Id,
		SHA1: image.Image.Sha1,
		SHA256: image.Image.Sha256,
	})
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := validateImageType(imageType); err != nil {
		return err
	}

	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("use --confirm to delete image")
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete image %s from %s/%s", imageID, locale, imageType)
		return nil
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

	err = edit.Images().Delete(client.GetPackageName(), edit.ID(), locale, imageType, imageID).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("Image deleted: %s", imageID)
	return nil
}

func runDeleteAll(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if err := validateImageType(imageType); err != nil {
		return err
	}

	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("use --confirm to delete all images of type '%s'", imageType)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete all images from %s/%s", locale, imageType)
		return nil
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

	_, err = edit.Images().Deleteall(client.GetPackageName(), edit.ID(), locale, imageType).Context(edit.Context()).Do()
	if err != nil {
		return err
	}

	if err := edit.Commit(); err != nil {
		return err
	}

	output.PrintSuccess("All images deleted for %s/%s", locale, imageType)
	return nil
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

	client, err := api.NewClient(cli.GetPackageName(), 5*time.Minute)
	if err != nil {
		return err
	}

	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	ctx := edit.Context()
	uploaded := 0

	// Walk directory: locale/imageType/files
	locales, err := os.ReadDir(absDir)
	if err != nil {
		return err
	}

	for _, localeEntry := range locales {
		if !localeEntry.IsDir() {
			continue
		}

		localeName := localeEntry.Name()
		localeDir := filepath.Join(absDir, localeName)

		types, err := os.ReadDir(localeDir)
		if err != nil {
			continue
		}

		for _, typeEntry := range types {
			if !typeEntry.IsDir() {
				continue
			}

			typeName := typeEntry.Name()
			if err := validateImageType(typeName); err != nil {
				continue
			}

			typeDir := filepath.Join(localeDir, typeName)
			files, err := os.ReadDir(typeDir)
			if err != nil {
				continue
			}

			for _, fileEntry := range files {
				if fileEntry.IsDir() {
					continue
				}

				filePath := filepath.Join(typeDir, fileEntry.Name())
				ext := strings.ToLower(filepath.Ext(fileEntry.Name()))
				if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
					continue
				}

				if cli.IsDryRun() {
					output.PrintInfo("Dry run: would upload %s to %s/%s", fileEntry.Name(), localeName, typeName)
					continue
				}

				file, err := os.Open(filePath)
				if err != nil {
					output.PrintWarning("Failed to open %s: %v", filePath, err)
					continue
				}

				_, err = edit.Images().Upload(client.GetPackageName(), edit.ID(), localeName, typeName).Media(file).Context(ctx).Do()
				file.Close()

				if err != nil {
					output.PrintWarning("Failed to upload %s: %v", fileEntry.Name(), err)
					continue
				}

				output.PrintInfo("Uploaded: %s/%s/%s", localeName, typeName, fileEntry.Name())
				uploaded++
			}
		}
	}

	if !cli.IsDryRun() && uploaded > 0 {
		if err := edit.Commit(); err != nil {
			return err
		}
	}

	output.PrintSuccess("Uploaded %d image(s)", uploaded)
	return nil
}
