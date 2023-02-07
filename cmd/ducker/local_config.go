package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// LocalConfig ducker local config structure
// Normally, this file is located at $DIR/.ducker.yaml
type LocalConfig struct {
	Run_Arg    []string
	Build_Arg  []string
	Mount_PWD  bool
  Default_Shell string
	LastExecID string
}

// Get default local config setting
func getDefaultLocalConfig() LocalConfig {
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
			"--build-arg UID=" + getTerminalCmdOut("id", "-u"),
			"--build-arg GID=" + getTerminalCmdOut("id", "-g"),
		},
    Default_Shell: "zsh",
		Mount_PWD: true,
	}

	return config
}

// Write default local config file
func writeDefaultLocalConfig() {
	configPath := getDefaultLocalConfigPath()
	config := getDefaultLocalConfig()
	config.Write(configPath)
}

// Get default local config path which is $PWD/.ducker.yaml
func getDefaultLocalConfigPath() string {
	homeDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	configPath := path.Join(homeDir, ".ducker.yaml")

	return configPath
}

// Read local config given the path
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

// Read default local config which is located at $PWD/.ducker.yaml
func readDefaultLocalConfig() LocalConfig {
	configPath := getDefaultLocalConfigPath()
	return readLocalConfig(configPath)
}

// IsEmpty returns true if run and build arg are empty
func (config LocalConfig) IsEmpty() bool {
	if len(config.Build_Arg) == 0 && len(config.Run_Arg) == 0 {
		return true
	}

	return false
}

// GetRunArg concatenates all run arguments
func (config LocalConfig) GetRunArg() string {
	return strings.Join(config.Run_Arg, " ")
}

// GetBuildArg concatenates all build arguments
func (config LocalConfig) GetBuildArg() string {
	return strings.Join(config.Build_Arg, " ")
}

// Write config file
func (config LocalConfig) Write(path string) {
	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal(err)
	}
	err2 := ioutil.WriteFile(path, data, 0644)

	if err2 != nil {
		log.Fatal(err2)
	}
}
