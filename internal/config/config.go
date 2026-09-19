// Package config loads user configuration for Logseq Doctor commands.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	appName        = "lqd"
	configFileName = "config.toml"
)

var errTidyUpConfigNotFound = errors.New("tidy-up config file not found")

// ForbiddenContentPolicy lists content that tidy-up reports as forbidden.
type ForbiddenContentPolicy struct {
	PageReferences []string `mapstructure:"page_references"`
	URLSubstrings  []string `mapstructure:"url_substrings"`
	TextSubstrings []string `mapstructure:"text_substrings"`
}

type fileConfig struct {
	TidyUp tidyUpConfig `mapstructure:"tidy-up"`
}

type tidyUpConfig struct {
	Forbidden ForbiddenContentPolicy `mapstructure:"forbidden"`
}

// TidyUpConfigPath returns the standard path for the tidy-up policy.
func TidyUpConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config directory: %w", err)
	}

	return filepath.Join(configDir, appName, configFileName), nil
}

// LoadTidyUpPolicy reads a tidy-up policy from path.
func LoadTidyUpPolicy(path string) (ForbiddenContentPolicy, error) {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ForbiddenContentPolicy{}, fmt.Errorf(
				"%w: %s; create it before running tidy-up", errTidyUpConfigNotFound, path,
			)
		}

		return ForbiddenContentPolicy{}, fmt.Errorf("read tidy-up config file %s: %w", path, err)
	}

	config := viper.New()
	config.SetConfigFile(path)
	config.SetConfigType("toml")

	err = config.ReadInConfig()
	if err != nil {
		return ForbiddenContentPolicy{}, fmt.Errorf("read tidy-up config file %s: %w", path, err)
	}

	var parsed fileConfig

	err = config.Unmarshal(&parsed)
	if err != nil {
		return ForbiddenContentPolicy{}, fmt.Errorf("parse tidy-up config file %s: %w", path, err)
	}

	return parsed.TidyUp.Forbidden, nil
}
