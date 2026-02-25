package devicetiers

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

// DeviceTiersCmd manages device tier configurations
var DeviceTiersCmd = &cobra.Command{
	Use:     "device-tiers",
	Aliases: []string{"dt"},
	Short:   "Manage device tier configurations",
	Long: `Manage device tier configurations for targeted content delivery.

Device tiers let you define groups of devices based on RAM,
device model, or system features, and deliver different
content to each tier.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List device tier configurations",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a device tier configuration",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a device tier configuration",
	RunE:  runCreate,
}

var (
	configID int64
	filePath string
)

func init() {
	getCmd.Flags().Int64Var(&configID, "config-id", 0, "device tier config ID")
	getCmd.MarkFlagRequired("config-id")

	createCmd.Flags().StringVar(&filePath, "file", "", "JSON file with device tier config")
	createCmd.MarkFlagRequired("file")

	DeviceTiersCmd.AddCommand(listCmd)
	DeviceTiersCmd.AddCommand(getCmd)
	DeviceTiersCmd.AddCommand(createCmd)
}

// DeviceTierInfo represents device tier config summary
type DeviceTierInfo struct {
	ConfigID   int64 `json:"config_id"`
	TierGroups int   `json:"tier_groups"`
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

	resp, err := client.Apps().DeviceTierConfigs.List(client.GetPackageName()).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to list device tier configs: %w", err)
	}

	if len(resp.DeviceTierConfigs) == 0 {
		output.PrintInfo("No device tier configurations found")
		return nil
	}

	result := make([]DeviceTierInfo, 0, len(resp.DeviceTierConfigs))
	for _, c := range resp.DeviceTierConfigs {
		groups := 0
		if c.DeviceGroups != nil {
			groups = len(c.DeviceGroups)
		}
		result = append(result, DeviceTierInfo{
			ConfigID:   c.DeviceTierConfigId,
			TierGroups: groups,
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

	config, err := client.Apps().DeviceTierConfigs.Get(
		client.GetPackageName(), configID,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get device tier config: %w", err)
	}

	return output.Print(config)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var config androidpublisher.DeviceTierConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create device tier configuration")
		return output.Print(config)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	created, err := client.Apps().DeviceTierConfigs.Create(
		client.GetPackageName(), &config,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create device tier config: %w", err)
	}

	output.PrintSuccess("Device tier configuration created: %d", created.DeviceTierConfigId)
	return output.Print(created)
}
