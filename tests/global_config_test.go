package duckerlib_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/jeikeilim/ducker/internal/duckerlib"
)

// TestGetDefaultGlobalConfig tests the GetDefaultGlobalConfig function
// from duckerlib.
func TestGetDefaultGlobalConfig(t *testing.T) {
	// Simulate user input for the actual GetDefaultGlobalConfig behavior
	input := "TestOrg\nTestName\nTestContact\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Write([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	w.Close()

	originalStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = originalStdin }()

	// Call the actual function from duckerlib
	// It expects empty strings to trigger prompts.
	actual := duckerlib.GetDefaultGlobalConfig("", "", "")

	expected := duckerlib.GlobalConfig{ // Prefixed with duckerlib.
		Organization: "TestOrg",
		Name:         "TestName",
		Contact:      "TestContact",
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("GetDefaultGlobalConfig(\"\", \"\", \"\") = %v, want %v", actual, expected)
	}
}

// TestGlobalConfig_IsEmpty tests the IsEmpty method of GlobalConfig
// from duckerlib.
func TestGlobalConfig_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		config   duckerlib.GlobalConfig // Prefixed with duckerlib.
		expected bool
	}{
		{
			name: "empty config",
			config: duckerlib.GlobalConfig{ // Prefixed with duckerlib.
				Organization: "",
				Name:         "",
				Contact:      "",
			},
			expected: true,
		},
		{
			name: "non-empty config",
			config: duckerlib.GlobalConfig{ // Prefixed with duckerlib.
				Organization: "TestOrg",
				Name:         "TestName",
				Contact:      "TestContact",
			},
			expected: false,
		},
		{
			name: "partially empty config",
			config: duckerlib.GlobalConfig{ // Prefixed with duckerlib.
				Organization: "TestOrg",
				Name:         "",
				Contact:      "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsEmpty(); got != tt.expected {
				t.Errorf("GlobalConfig.IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGlobalConfig_Write_and_Read tests the Write and ReadGlobalConfig functions
// from duckerlib (which use YAML).
func TestGlobalConfig_Write_and_Read(t *testing.T) {
	config := duckerlib.GlobalConfig{ // Prefixed with duckerlib.
		Organization: "TestOrgYAML",
		Name:         "TestNameYAML",
		Contact:      "TestContactYAML",
	}

	tmpfile, err := ioutil.TempFile("", "global_config_test_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up

	// Write the config using the actual Write method (panics on error via log.Fatal)
	// We need to close the file before writing, as Write will try to open it.
	tmpfilePath := tmpfile.Name()
	tmpfile.Close()


	// The actual Write method uses log.Fatal on error.
	// If an error occurs, the test will terminate here.
	config.Write(tmpfilePath)


	// Read the config using the actual readGlobalConfig (panics on error via log.Fatal)
	// If an error occurs, the test will terminate here.
	readConf := duckerlib.ReadGlobalConfig(tmpfilePath) // Prefixed with duckerlib.

	// Since ReadGlobalConfig prints to stdout, we might want to suppress that for tests
	// or ignore it. For this refactor, we'll leave it as is.

	if !reflect.DeepEqual(config, readConf) {
		t.Errorf("ReadGlobalConfig() = %v, want %v", readConf, config)
	}
}

// Note: Duplicated struct/function removals were done in a previous refactoring step
// when source files were moved into the tests/ directory and package was main.
// This step focuses on changing package to _test and using the imported duckerlib.
