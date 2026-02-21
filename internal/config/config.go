package config

import (
	"os"
	"fmt"
	"encoding/json"
)

const configName = ".gatorconfig.json"

type Config struct {
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (config *Config) SetUser(username string) error {
	config.CurrentUserName = username
	err := writeConfig(config)
	return err
}

func Read() (Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (config *Config) Print() {
	fmt.Printf("Current %s:\n", configName)
	fmt.Printf("%#v\n", config)
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := homeDir + "/" + configName
	return path, nil
}

func writeConfig(config *Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	configAsBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, configAsBytes, os.ModePerm) 
	return err
}
