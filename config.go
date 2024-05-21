package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	BaseURL string `json:"base_url"`
	UserID  int    `json:"user_id"`
	APIKey  string `json:"api_key"`
}

func ReadConfig() (*Config, error) {
	paths := []string{
		"./config.json",
		"$HOME/.config/lazyop/config.json",
	}
	var config Config
	for _, path := range paths {
		expandedPath := os.ExpandEnv(path)
		if fileExists(expandedPath) {
			data, err := os.ReadFile(expandedPath)
			if err != nil {
				return nil, fmt.Errorf("error reading file %s: %v", expandedPath, err)
			}
			if err := json.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("error parsing file %s: %v", expandedPath, err)
			}
			return &config, nil
		}
	}
	return nil, fmt.Errorf("no config file found")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
