package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/AndroidPoet/playconsole-cli/internal/config"
)

func TestGetPackageNameFallsBackToProfileDefault(t *testing.T) {
	previousPackageName := packageName
	previousTimeout := timeout
	packageName = ""
	timeout = ""
	viper.Set("package", "")
	t.Cleanup(func() {
		packageName = previousPackageName
		timeout = previousTimeout
		viper.Set("package", "")
	})

	configFile := filepath.Join(t.TempDir(), "config.json")
	configJSON := `{
  "default_profile": "default",
  "profiles": {
    "default": {
      "name": "default",
      "credentials_path": "/tmp/creds.json",
      "default_package": "com.example.profile"
    }
  }
}`
	if err := os.WriteFile(configFile, []byte(configJSON), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if err := config.Init(configFile, ""); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	if got := GetPackageName(); got != "com.example.profile" {
		t.Fatalf("expected default package from profile, got %q", got)
	}
}

func TestResolveTimeoutUsesFallbackWithoutExplicitOverride(t *testing.T) {
	previousTimeout := timeout
	timeout = ""
	t.Cleanup(func() {
		timeout = previousTimeout
	})

	fallback := 5 * time.Minute
	if got := ResolveTimeout(fallback); got != fallback {
		t.Fatalf("expected fallback timeout %s, got %s", fallback, got)
	}
}

func TestResolveTimeoutUsesExplicitOverride(t *testing.T) {
	previousTimeout := timeout
	timeout = "90s"
	t.Cleanup(func() {
		timeout = previousTimeout
	})

	if got := ResolveTimeout(5 * time.Minute); got != 90*time.Second {
		t.Fatalf("expected explicit timeout override, got %s", got)
	}
}
