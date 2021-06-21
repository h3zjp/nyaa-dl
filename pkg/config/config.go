package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents a configuration of downloading.
type Config struct {
	Targets []Target `json:"targets"`
}

// Target represents what should a client download.
type Target struct {
	// Title represents a free title of this target.
	Title string `json:"title"`
	// Category of nyaa.
	Category string `json:"category"`
	// Query is a search keyword.
	Query string `json:"query"`
	// RequiredDownloads represents the minimum num of downloads to download.
	RequiredDownloads int `json:"requiredDownloads"`
	// MaxPage to crawl nyaa.
	MaxPage int `json:"maxPage"`
}

// ReadConfig reads a configuration file.
func ReadConfig(filePath string) (*Config, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read a file: %w", err)
	}

	var conf *Config
	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse a config: %w", err)
	}

	return conf, nil
}
