package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Profile represents an authentication profile
type Profile struct {
	Name            string `json:"name"`
	CredentialsPath string `json:"credentials_path,omitempty"`
	CredentialsB64  string `json:"credentials_b64,omitempty"`
	DefaultPackage  string `json:"default_package,omitempty"`
}

// Config represents the playconsole-cli configuration
type Config struct {
	DefaultProfile string             `json:"default_profile"`
	Profiles       map[string]Profile `json:"profiles"`
}

var (
	cfg          *Config
	currentProfile *Profile
	debugMode    bool
	configPath   string
)

// Init initializes the configuration
func Init(cfgFile, profileName string) error {
	// Determine config path
	if cfgFile != "" {
		configPath = cfgFile
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, ".playconsole-cli", "config.json")
	}

	// Load or create config
	cfg = &Config{
		DefaultProfile: "default",
		Profiles:       make(map[string]Profile),
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Determine which profile to use
	if profileName == "" {
		profileName = viper.GetString("profile")
	}
	if profileName == "" {
		profileName = cfg.DefaultProfile
	}
	if profileName == "" {
		profileName = "default"
	}

	// Load profile
	if p, ok := cfg.Profiles[profileName]; ok {
		currentProfile = &p
	} else {
		// Try to create profile from environment
		currentProfile = &Profile{Name: profileName}
	}

	// Override credentials from environment if set
	if credPath := viper.GetString("credentials_path"); credPath != "" {
		currentProfile.CredentialsPath = credPath
	}
	if credB64 := viper.GetString("credentials_b64"); credB64 != "" {
		currentProfile.CredentialsB64 = credB64
	}

	return nil
}

// Save saves the current configuration
func Save() error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	return cfg
}

// GetProfile returns the current profile
func GetProfile() *Profile {
	return currentProfile
}

// SetProfile sets a profile in the configuration
func SetProfile(p Profile) {
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	cfg.Profiles[p.Name] = p
}

// DeleteProfile removes a profile from the configuration
func DeleteProfile(name string) {
	if cfg.Profiles != nil {
		delete(cfg.Profiles, name)
	}
}

// SetDefaultProfile sets the default profile name
func SetDefaultProfile(name string) {
	cfg.DefaultProfile = name
}

// GetCredentials returns the service account credentials JSON
func GetCredentials() ([]byte, error) {
	if currentProfile == nil {
		return nil, fmt.Errorf("no profile configured. Run 'playconsole-cli auth login' first")
	}

	// Try base64 encoded credentials first
	if currentProfile.CredentialsB64 != "" {
		data, err := base64.StdEncoding.DecodeString(currentProfile.CredentialsB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode credentials: %w", err)
		}
		return data, nil
	}

	// Try credentials file
	if currentProfile.CredentialsPath != "" {
		data, err := os.ReadFile(currentProfile.CredentialsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials file: %w", err)
		}
		return data, nil
	}

	return nil, fmt.Errorf("no credentials configured for profile '%s'. Run 'playconsole-cli auth login' first", currentProfile.Name)
}

// SetDebug sets debug mode
func SetDebug(d bool) {
	debugMode = d
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	return debugMode
}

// GetConfigPath returns the config file path
func GetConfigPath() string {
	return configPath
}

// ListProfiles returns all profile names
func ListProfiles() []string {
	if cfg == nil || cfg.Profiles == nil {
		return nil
	}
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	return names
}
