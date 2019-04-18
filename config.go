package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

const (
	configFileName = "outbot.json"
)

// Config for OutBot.
type Config struct {
	Guild  `json:"guild"`
	Sheets `json:"sheets"`
}

// Guild config contains information about the discord guild.
type Guild struct {
	ID string `json:"guildId"`
	// TODO: Add channel parameters
}

type Sheets struct {
	AuthCode string `json:"authCode"`
}

// readConfig in the dir.
// If one does not exist at the directory then a placeholder will be created.
func readConfig(dir string) Config {
	path := filepath.Join(dir, configFileName)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		createConfigPlaceholder(path)
		panic(fmt.Sprintf("Unable to read config file: %v", err)) // Do exit instead with a friendly message to edit config
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		panic(fmt.Sprintf("Unable to parse config: %v", err))
	}

	return config
}

func createConfigPlaceholder(path string) {
	placeholder, err := json.MarshalIndent(Config{}, "", "  ")
	if err != nil {
		fmt.Println("Unable to marshal new config:", err.Error())
		return
	}

	err = ioutil.WriteFile(path, placeholder, 644)
	if err != nil {
		fmt.Println("Failed to write config placeholder:", err.Error())
		return
	}
}
