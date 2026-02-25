package deobfuscation

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// DeobfuscationCmd manages deobfuscation/mapping file uploads
var DeobfuscationCmd = &cobra.Command{
	Use:   "deobfuscation",
	Short: "Manage deobfuscation (mapping) files",
	Long: `Upload ProGuard/R8 mapping files or native debug symbols
to enable readable stack traces in crash reports.

This is essential for debugging production crashes when your
app uses code shrinking or obfuscation.`,
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a deobfuscation file",
	Long: `Upload a ProGuard/R8 mapping file or native debug symbols ZIP.

Types:
  proguard     - ProGuard/R8 mapping.txt file
  native-code  - Native debug symbols (symbols.zip)`,
	RunE: runUpload,
}

var (
	versionCode int64
	filePath    string
	fileType    string
)

func init() {
	uploadCmd.Flags().Int64Var(&versionCode, "version-code", 0, "APK/AAB version code")
	uploadCmd.Flags().StringVar(&filePath, "file", "", "path to mapping file")
	uploadCmd.Flags().StringVar(&fileType, "type", "proguard", "file type: proguard, native-code")
	uploadCmd.MarkFlagRequired("version-code")
	uploadCmd.MarkFlagRequired("file")

	DeobfuscationCmd.AddCommand(uploadCmd)
}

// UploadResult represents the upload result
type UploadResult struct {
	VersionCode int64  `json:"version_code"`
	FileType    string `json:"file_type"`
	Status      string `json:"status"`
}

func runUpload(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	// Map user-friendly type to API type
	apiType := mapFileType(fileType)
	if apiType == "" {
		return fmt.Errorf("invalid type '%s': use 'proguard' or 'native-code'", fileType)
	}

	// Verify file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would upload %s (%d bytes) as %s for version %d",
			filePath, info.Size(), apiType, versionCode)
		return nil
	}

	// Use longer timeout for file uploads
	client, err := api.NewClient(cli.GetPackageName(), 5*time.Minute)
	if err != nil {
		return err
	}

	// Create edit session
	edit, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer edit.Close()

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload
	resp, err := edit.DeobfuscationFiles().Upload(
		client.GetPackageName(), edit.ID(), versionCode, apiType,
	).Media(file).Context(edit.Context()).Do()
	if err != nil {
		return fmt.Errorf("failed to upload deobfuscation file: %w", err)
	}

	// Commit edit
	if err := edit.Commit(); err != nil {
		return err
	}

	result := UploadResult{
		VersionCode: versionCode,
		FileType:    resp.DeobfuscationFile.SymbolType,
		Status:      "uploaded",
	}

	output.PrintSuccess("Deobfuscation file uploaded for version %d", versionCode)
	return output.Print(result)
}

func mapFileType(t string) string {
	switch strings.ToLower(t) {
	case "proguard":
		return "proguard"
	case "native-code", "nativecode", "native":
		return "nativeCode"
	default:
		return ""
	}
}
