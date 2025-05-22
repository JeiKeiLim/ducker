package duckerlib

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

// GetDefaultLocalConfig default local config setting
func GetDefaultLocalConfig() LocalConfig {
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
		},
    Default_Shell: "zsh",
		Mount_PWD: true,
	}

	return config
}

// WriteDefaultLocalConfig default local config file
func WriteDefaultLocalConfig() {
	configPath := GetDefaultLocalConfigPath()
	config := GetDefaultLocalConfig()
	config.Write(configPath)
}

// GetDefaultLocalConfigPath default local config path which is $PWD/.ducker.yaml
func GetDefaultLocalConfigPath() string {
	homeDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	configPath := path.Join(homeDir, ".ducker.yaml")

	return configPath
}

// ReadLocalConfig local config given the path
func ReadLocalConfig(path string) LocalConfig {
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

// ReadDefaultLocalConfig default local config which is located at $PWD/.ducker.yaml
func ReadDefaultLocalConfig() LocalConfig {
	configPath := GetDefaultLocalConfigPath()
	return ReadLocalConfig(configPath)
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
