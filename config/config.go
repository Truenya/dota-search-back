package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func GetToken() string {
	token := os.Getenv("VK_TOKEN")
	if token != "" {
		return token
	}

	if token, err := GetTokenFromFile(); err == nil && token != "" {
		return token
	}

	panic("VK_TOKEN not found, set it in env or in file. U can get it from https://vkhost.github.io/")
}

func GetTokenFromFile() (string, error) {
	const configFile = "config.json"

	file, err := os.OpenFile(configFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", configFile, err)
	}

	var config struct {
		Token string
	}

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return "", fmt.Errorf(" decode %s: %w", configFile, err)
	}

	return config.Token, nil
}

func GetIP() string {
	ip := os.Getenv("IP")
	if ip != "" {
		return ip
	}

	return "127.0.0.1"
}
