package duckerlib_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jeikeilim/ducker/internal/duckerlib"
)

// Test cases start here

func TestGetDefaultLocalConfig(t *testing.T) {
	// Expected values based on the actual GetDefaultLocalConfig in duckerlib
	expected := duckerlib.LocalConfig{ // Prefixed with duckerlib.
		Run_Arg: []string{
			"--privileged",
			"-e DISPLAY=" + os.Getenv("DISPLAY"), // Relies on environment variable
			"-e TERM=xterm-256color",
			"-v /tmp/.X11-unix:/tmp/.X11-unix:ro",
			"-v /dev:/dev",
			"--network host",
		},
		Build_Arg:     []string{},
		Mount_PWD:     true,
		Default_Shell: "zsh",
		// LastExecID is not set by GetDefaultLocalConfig, so it will be its zero value ""
	}
	actual := duckerlib.GetDefaultLocalConfig() // Prefixed with duckerlib.

	// For Run_Arg, os.Getenv("DISPLAY") could be "" in test environment.
	// Let's construct the expected DISPLAY arg based on the actual environment
	expected.Run_Arg[1] = "-e DISPLAY=" + os.Getenv("DISPLAY")

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("GetDefaultLocalConfig() = \n%#v, \nwant \n%#v", actual, expected)
	}
}

func TestLocalConfig_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		config   duckerlib.LocalConfig // Prefixed with duckerlib.
		expected bool
	}{
		{
			name: "empty args",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Run_Arg:   []string{},
				Build_Arg: []string{},
			},
			expected: true,
		},
		{
			name: "non-empty Run_Arg",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Run_Arg:   []string{"-it"},
				Build_Arg: []string{},
			},
			expected: false,
		},
		{
			name: "non-empty Build_Arg",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Run_Arg:   []string{},
				Build_Arg: []string{"--no-cache"},
			},
			expected: false,
		},
		{
			name: "non-empty both args",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Run_Arg:   []string{"-it"},
				Build_Arg: []string{"--no-cache"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsEmpty(); got != tt.expected {
				t.Errorf("LocalConfig.IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLocalConfig_GetRunArg(t *testing.T) {
	tests := []struct {
		name     string
		config   duckerlib.LocalConfig // Prefixed with duckerlib.
		expected string
	}{
		{
			name: "sample Run_Arg",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Run_Arg: []string{"-it", "--rm", "-p", "8080:80"},
			},
			expected: "-it --rm -p 8080:80",
		},
		{
			name: "empty Run_Arg",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Run_Arg: []string{},
			},
			expected: "",
		},
		{
			name: "single item Run_Arg",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Run_Arg: []string{"-d"},
			},
			expected: "-d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetRunArg(); got != tt.expected {
				t.Errorf("LocalConfig.GetRunArg() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLocalConfig_GetBuildArg(t *testing.T) {
	tests := []struct {
		name     string
		config   duckerlib.LocalConfig // Prefixed with duckerlib.
		expected string
	}{
		{
			name: "sample Build_Arg",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Build_Arg: []string{"--no-cache", "--pull"},
			},
			expected: "--no-cache --pull",
		},
		{
			name: "empty Build_Arg",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Build_Arg: []string{},
			},
			expected: "",
		},
		{
			name: "single item Build_Arg",
			config: duckerlib.LocalConfig{ // Prefixed with duckerlib.
				Build_Arg: []string{"--compress"},
			},
			expected: "--compress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetBuildArg(); got != tt.expected {
				t.Errorf("LocalConfig.GetBuildArg() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLocalConfig_Write_and_Read(t *testing.T) {
	originalConfig := duckerlib.LocalConfig{ // Prefixed with duckerlib.
		Run_Arg:       []string{"-d", "--name", "my-container-yaml"},
		Build_Arg:     []string{"--tag", "latest-yaml"},
		Mount_PWD:     false,
		Default_Shell: "/bin/sh",
		LastExecID:    "some-id", // Actual LocalConfig has this field
	}

	tmpfile, err := ioutil.TempFile("", "local_config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	filePath := tmpfile.Name()
	tmpfile.Close() // Close before Write, as it opens the file itself.
	defer os.Remove(filePath) // Clean up

	// Write the config using the actual Write method (from duckerlib)
	// This method uses log.Fatal on error, so test will terminate if it fails.
	originalConfig.Write(filePath)

	// Read the config using the actual ReadLocalConfig method (from duckerlib)
	// This method also uses log.Fatal on error.
	readConf := duckerlib.ReadLocalConfig(filePath) // Prefixed with duckerlib.

	if !reflect.DeepEqual(originalConfig, readConf) {
		t.Errorf("Read config = \n%#v, \nwant \n%#v", readConf, originalConfig)
	}
}
