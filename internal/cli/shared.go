package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/AndroidPoet/playconsole-cli/internal/config"
)

var (
	packageName string
	profile     string
	timeout     string
	dryRun      bool
)

// SetPackageName sets the package name from the flag
func SetPackageName(p string) {
	packageName = p
}

// SetProfile sets the profile from the flag
func SetProfile(p string) {
	profile = p
}

// SetTimeout sets the timeout from the flag
func SetTimeout(t string) {
	timeout = t
}

// SetDryRun sets the dry-run mode
func SetDryRun(d bool) {
	dryRun = d
}

// GetPackageName returns the package name from flag, env, or config
func GetPackageName() string {
	if packageName != "" {
		return packageName
	}
	if pkg := viper.GetString("package"); pkg != "" {
		return pkg
	}
	if profile := config.GetProfile(); profile != nil {
		return profile.DefaultPackage
	}
	return ""
}

// GetProfile returns the profile name from flag, env, or config
func GetProfile() string {
	if profile != "" {
		return profile
	}
	p := viper.GetString("profile")
	if p == "" {
		return "default"
	}
	return p
}

// GetTimeout returns the timeout duration string
func GetTimeout() string {
	if timeout != "" {
		return timeout
	}
	t := viper.GetString("timeout")
	if t == "" {
		return "60s"
	}
	return t
}

// ResolveTimeout returns an explicit user timeout when present, otherwise fallback.
func ResolveTimeout(fallback time.Duration) time.Duration {
	if timeout == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(timeout)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

// IsDryRun returns whether dry-run mode is enabled
func IsDryRun() bool {
	return dryRun
}

// RequirePackage validates package name is set
func RequirePackage(cmd *cobra.Command) error {
	pkg := GetPackageName()
	if pkg == "" {
		return fmt.Errorf("package name required: use --package flag or set GPC_PACKAGE environment variable")
	}
	return nil
}

// CheckConfirm validates confirmation for destructive operations
func CheckConfirm(cmd *cobra.Command) error {
	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("this is a destructive operation. Use --confirm to proceed")
	}
	return nil
}

// MustMarkFlagRequired marks a flag as required during command initialization.
func MustMarkFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Sprintf("failed to mark flag %q as required on %q: %v", name, cmd.Name(), err))
	}
}

// MustMarkPersistentFlagRequired marks a persistent flag as required during command initialization.
func MustMarkPersistentFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkPersistentFlagRequired(name); err != nil {
		panic(fmt.Sprintf("failed to mark persistent flag %q as required on %q: %v", name, cmd.Name(), err))
	}
}
