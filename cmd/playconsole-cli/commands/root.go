package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/initcmd"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/config"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var (
	cfgFile     string
	packageName string
	profile     string
	outputFmt   string
	prettyPrint bool
	quiet       bool
	debug       bool
	timeout     string
	dryRun      bool

	versionStr string
	commitStr  string
	dateStr    string
)

var rootCmd = &cobra.Command{
	Use:   "playconsole-cli",
	Short: "Google Play Console CLI",
	Long: `playconsole-cli is a fast, lightweight, and scriptable CLI for Google Play Console.

It provides comprehensive automation for Android app publishing workflows,
designed for CI/CD pipelines and developer productivity.

Design Philosophy:
  • JSON-first output for automation
  • Explicit flags over cryptic shortcuts
  • No interactive prompts
  • Clean non-zero exits on failure`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip for completion and help commands
		if cmd.Name() == "completion" || cmd.Name() == "help" || cmd.Name() == "__complete" {
			return nil
		}

		// Load .gpc.yaml project config if it exists
		if cwd, err := os.Getwd(); err == nil {
			if projectCfg := initcmd.FindProjectConfig(cwd); projectCfg != "" {
				viper.SetConfigFile(projectCfg)
				viper.SetConfigType("yaml")
				_ = viper.MergeInConfig()
			}
		}

		// Sync flags to cli package
		cli.SetPackageName(packageName)
		cli.SetProfile(profile)
		if cmd.Flags().Changed("timeout") || os.Getenv("GPC_TIMEOUT") != "" {
			cli.SetTimeout(timeout)
		} else {
			cli.SetTimeout("")
		}
		cli.SetDryRun(dryRun)

		// Initialize config
		if err := config.Init(cfgFile, profile); err != nil {
			return err
		}

		// Setup output formatter
		output.Setup(outputFmt, prettyPrint, quiet)

		// Set debug mode
		if debug {
			config.SetDebug(true)
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func SetVersionInfo(version, commit, date string) {
	versionStr = version
	commitStr = commit
	dateStr = date
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.playconsole-cli/config.json)")
	rootCmd.PersistentFlags().StringVarP(&packageName, "package", "p", "", "app package name (or GPC_PACKAGE env)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "auth profile name (or GPC_PROFILE env)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "json", "output format: json, table, minimal, tsv, csv, yaml, markdown")
	rootCmd.PersistentFlags().BoolVar(&prettyPrint, "pretty", false, "pretty-print JSON output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "show API requests/responses")
	rootCmd.PersistentFlags().StringVar(&timeout, "timeout", "60s", "request timeout")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview changes without applying")

	// Bind to viper
	mustBindPFlag("package", rootCmd.PersistentFlags().Lookup("package"))
	mustBindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	mustBindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	mustBindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	mustBindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))

	// Environment variable bindings
	mustBindEnv("package", "GPC_PACKAGE")
	mustBindEnv("profile", "GPC_PROFILE")
	mustBindEnv("output", "GPC_OUTPUT")
	mustBindEnv("debug", "GPC_DEBUG")
	mustBindEnv("timeout", "GPC_TIMEOUT")
	mustBindEnv("credentials_path", "GPC_CREDENTIALS_PATH")
	mustBindEnv("credentials_b64", "GPC_CREDENTIALS_B64")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("playconsole-cli %s\n", versionStr)
			fmt.Printf("  commit: %s\n", commitStr)
			fmt.Printf("  built:  %s\n", dateStr)
		},
	})
}

// GetRootCmd returns the root command for adding subcommands
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func mustBindPFlag(key string, flag *pflag.Flag) {
	if err := viper.BindPFlag(key, flag); err != nil {
		panic(fmt.Sprintf("failed to bind flag %q: %v", key, err))
	}
}

func mustBindEnv(key string, envs ...string) {
	args := append([]string{key}, envs...)
	if err := viper.BindEnv(args...); err != nil {
		panic(fmt.Sprintf("failed to bind env for %q: %v", key, err))
	}
}

// ExitError exits with error
func ExitError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
