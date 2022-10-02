package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type LocalConfig struct {
	Run_Arg   []string
	Build_Arg []string
	Mount_PWD bool
}

type GlobalConfig struct {
	Organization string
	Name         string
	Contact      string
}

func getDefaultLocalConfig() string {
	homeDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	configPath := path.Join(homeDir, ".ducker.yaml")

	return configPath
}

func writeDefaultLocalConfig() {
	configPath := getDefaultLocalConfig()
	writeLocalConfig(configPath)
}

func writeLocalConfig(path string) {
	config := LocalConfig{
		Run_Arg: []string{
			"--privileged",
			"-e DISPLAY=" + os.Getenv("DISPLAY"),
			"-e TERM=xterm-256color",
			"-v /tmp/.X11-unix:/tmp/.X11-unix:ro",
			"-v /dev:/dev",
			"--network host",
		},
		Build_Arg: []string{
			"--build-arg UID=" + runTerminalCmd("id", "-u"),
			"--build-arg GID=" + runTerminalCmd("id", "-g"),
		},
		Mount_PWD: true,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal(err)
	}

	err2 := ioutil.WriteFile(path, data, 0644)

	if err2 != nil {
		log.Fatal(err2)
	}
}

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
	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal(err)
	}
	err2 := ioutil.WriteFile(configPath, data, 0644)

	if err2 != nil {
		log.Fatal(err2)
	}

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

func (config GlobalConfig) IsEmpty() bool {
	if config.Organization == "" && config.Contact == "" && config.Name == "" {
		return true
	}

	return false
}

func (config GlobalConfig) Write() {
	writeDefaultGlobalConfig(config.Organization, config.Name, config.Contact)
}

func readLocalConfig(path string) LocalConfig {
	yfile, err := ioutil.ReadFile(path)
	if err != nil {
		return LocalConfig{}
	}

	config := LocalConfig{}

	err2 := yaml.Unmarshal([]byte(yfile), &config)
	if err2 != nil {
		log.Fatal(err2)
	}

	return config
}

func readDefaultLocalConfig() LocalConfig {
	configPath := getDefaultLocalConfig()
	return readLocalConfig(configPath)
}

func (config LocalConfig) IsEmpty() bool {
	if len(config.Build_Arg) == 0 && len(config.Run_Arg) == 0 {
		return true
	}

	return false
}

func (config LocalConfig) GetRunArg() string {
	return strings.Join(config.Run_Arg, " ")
}

func (config LocalConfig) GetBuildArg() string {
	return strings.Join(config.Build_Arg, " ")
}
