package initcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// InitCmd initializes a project configuration file
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project configuration",
	Long: `Create a .gpc.yaml configuration file in the current directory.

This file stores project-level defaults (package name, default track,
output format) so you don't need to pass them as flags every time.

The CLI automatically detects .gpc.yaml in the current directory
or any parent directory.`,
	RunE: runInit,
}

var (
	initPackage string
	initTrack   string
	initOutput  string
	force       bool
)

func init() {
	InitCmd.Flags().StringVar(&initPackage, "package", "", "app package name (e.g., com.example.app)")
	InitCmd.Flags().StringVar(&initTrack, "track", "internal", "default release track")
	InitCmd.Flags().StringVar(&initOutput, "output", "json", "default output format")
	InitCmd.Flags().BoolVar(&force, "force", false, "overwrite existing .gpc.yaml")
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	configPath := filepath.Join(cwd, ".gpc.yaml")

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf(".gpc.yaml already exists (use --force to overwrite)")
	}

	// Build config content
	content := "# playconsole-cli project configuration\n"
	content += "# See: https://github.com/AndroidPoet/playconsole-cli\n\n"

	if initPackage != "" {
		content += fmt.Sprintf("package: %s\n", initPackage)
	} else {
		content += "# package: com.example.app\n"
	}

	content += fmt.Sprintf("track: %s\n", initTrack)
	content += fmt.Sprintf("output: %s\n", initOutput)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .gpc.yaml: %w", err)
	}

	output.PrintSuccess("Created %s", configPath)
	return nil
}

// FindProjectConfig walks up from dir to find .gpc.yaml
func FindProjectConfig(dir string) string {
	for {
		candidate := filepath.Join(dir, ".gpc.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
