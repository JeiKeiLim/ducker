package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// GlobalConfig ducker global config structure
// Normally, this file is located at $HOME/.ducker_global.yaml
type GlobalConfig struct {
	Organization string
	Name         string
	Contact      string
}

// Get default global config
// Asks config values if "" has given in the arguments
func getDefaultGlobalConfig(organization string, name string, contact string) GlobalConfig {
	if organization == "" {
		organization = "jeikeilim"

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter your organization name(Default: %s): ", organization)
		keyIn, err := reader.ReadString('\n')
		checkError(err)
		if strings.TrimSpace(keyIn) != "" {
			organization = keyIn
		}
		organization = strings.TrimSpace(organization)
	}

	if name == "" {
		name = "Anonymous"
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Enter your name(Default: %s): ", name)
		keyIn, err := reader.ReadString('\n')
		checkError(err)
		if strings.TrimSpace(keyIn) != "" {
			name = keyIn
		}
		name = strings.TrimSpace(name)
	}

	if contact == "" {
		contact = "None"
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Enter contact address (Default: %s): ", contact)
		keyIn, err := reader.ReadString('\n')
		checkError(err)
		if strings.TrimSpace(keyIn) != "" {
			contact = keyIn
		}
		contact = strings.TrimSpace(contact)
	}

	globalConfig := GlobalConfig{
		Organization: organization,
		Name:         name,
		Contact:      contact,
	}

	return globalConfig
}

// Get default global config path
func getDefaultGlobalConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to fetch user home directory! ", err)
	}
	configPath := path.Join(homeDir, ".ducker_global.yaml")

	return configPath
}

func writeDefaultGlobalConfig(organization string, name string, contact string) {
	configPath := getDefaultGlobalConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		return
	}

	config := GlobalConfig{
		Organization: organization,
		Name:         name,
		Contact:      contact,
	}
	config.Write(configPath)

	fmt.Println("Global configuration has been written")
	fmt.Println("I won't be asking this again, please look at", configPath, "instaed.")
}

func readGlobalConfig(path string) GlobalConfig {
	yfile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	config := GlobalConfig{}

	err2 := yaml.Unmarshal([]byte(yfile), &config)
	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println("Global config has been read from", path)
	return config
}

func readDefaultGlobalConfig() GlobalConfig {
	configPath := getDefaultGlobalConfigPath()
	if _, err := os.Stat(configPath); err != nil {
		return GlobalConfig{}
	}

	return readGlobalConfig(configPath)
}

// IsEmpty returns true if all member variable strings are ""
func (config GlobalConfig) IsEmpty() bool {
	if config.Organization == "" && config.Contact == "" && config.Name == "" {
		return true
	}

	return false
}

func (config GlobalConfig) Write(path string) {
	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal(err)
	}
	err2 := ioutil.WriteFile(path, data, 0644)

	if err2 != nil {
		log.Fatal(err2)
	}
}
