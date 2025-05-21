package duckerlib_test

import (
	"bufio" // Still needed for TestAsksAreYouSure's os.Pipe interaction if not directly using duckerlib.AsksAreYouSure's reader
	"bytes" // Still needed for TestAsksAreYouSure
	"fmt"   // Still needed for httptest server and error messages
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jeikeilim/ducker/internal/duckerlib"
)

// --- Tests for funcs in duckerlib/utils.go ---

func TestGetTerminalCmdOut(t *testing.T) {
	expectedOutput := "test_output"
	actualOutput := duckerlib.GetTerminalCmdOut("echo", expectedOutput) // Prefixed

	if actualOutput != expectedOutput {
		t.Errorf("duckerlib.GetTerminalCmdOut(\"echo\", %q) = %q, want %q", expectedOutput, actualOutput, expectedOutput)
	}

	// Regarding the non-existent command test:
	// The actual duckerlib.GetTerminalCmdOut calls duckerlib.CheckError, which os.Exits.
	// A robust test for this would involve running the call in a separate process
	// and checking its exit code. This is beyond typical unit test scope.
	// If we call it directly and it exits, the test suite itself might terminate.
	// For now, this part of the test is implicitly testing that CheckError is called
	// if the command fails, though it doesn't assert the exit.
	// A command like "very_unlikely_command_to_exist_anywhere" might not error out from exec.Run
	// but return empty, in which case the original assertion holds.
	// If it does error, os.Exit happens.
	nonExistentCmdOutput := duckerlib.GetTerminalCmdOut("very_unlikely_command_to_exist_anywhere", "") // Prefixed
	if nonExistentCmdOutput != "" {
		t.Errorf("duckerlib.GetTerminalCmdOut(\"non_existent_command\", \"\") = %q, want \"\"", nonExistentCmdOutput)
	}
}

func TestGetArchType(t *testing.T) {
	arch := duckerlib.GetArchType() // Prefixed
	if arch == "" {
		t.Errorf("duckerlib.GetArchType() returned an empty string. Output of 'uname -m' might be empty or command failed.")
	}
	t.Logf("Detected architecture via duckerlib.GetArchType (uname -m): %s", arch)
}

func TestWriteFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test_write_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	filePath := tmpfile.Name()
	tmpfile.Close()
	defer os.Remove(filePath)

	content1 := "Hello, World!"
	content2 := "New Content Here"

	if !duckerlib.WriteFile(content1, filePath, false) { // Prefixed
		t.Errorf("duckerlib.WriteFile(content1, path, false) returned false, want true for initial write")
	}
	data, _ := ioutil.ReadFile(filePath)
	if string(data) != content1 {
		t.Errorf("File content after initial write = %q, want %q", string(data), content1)
	}

	if duckerlib.WriteFile(content2, filePath, false) { // Prefixed
		t.Errorf("duckerlib.WriteFile(content2, path, false) returned true, want false for existing file")
	}
	data, _ = ioutil.ReadFile(filePath)
	if string(data) != content1 {
		t.Errorf("File content after failed overwrite (overwrite=false) = %q, want %q", string(data), content1)
	}

	if !duckerlib.WriteFile(content2, filePath, true) { // Prefixed
		t.Errorf("duckerlib.WriteFile(content2, path, true) returned false, want true for overwrite")
	}
	data, _ = ioutil.ReadFile(filePath)
	if string(data) != content2 {
		t.Errorf("File content after successful overwrite (overwrite=true) = %q, want %q", string(data), content2)
	}
}

func TestGetContentFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/success" {
			fmt.Fprintln(w, "hello world")
		} else if r.URL.Path == "/error" {
			http.Error(w, "server error", http.StatusInternalServerError)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	content := duckerlib.GetContentFromURL(server.URL + "/success") // Prefixed
	expectedContent := "hello world"
	if strings.TrimSpace(content) != expectedContent {
		t.Errorf("duckerlib.GetContentFromURL (success) = %q, want %q", content, expectedContent)
	}

	content = duckerlib.GetContentFromURL(server.URL + "/error") // Prefixed
	if content != "" {
		t.Errorf("duckerlib.GetContentFromURL (server error) = %q, want \"\"", content)
	}
	
	content = duckerlib.GetContentFromURL(server.URL + "/notfoundpath") // Prefixed
	if content != "" {
		t.Errorf("duckerlib.GetContentFromURL (server 404) = %q, want \"\"", content)
	}

	content = duckerlib.GetContentFromURL("http://localhost:12346/nonexistent") // Prefixed
	if content != "" {
		t.Errorf("duckerlib.GetContentFromURL (non-existent server) = %q, want \"\"", content)
	}
}

func TestAsksAreYouSure(t *testing.T) {
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	tests := []struct {
		name        string
		input       string
		prompt      string
		expected    bool
	}{
		{"Input y", "y\n", "Test Y", true},
		{"Input n", "n\n", "Test N", false},
		{"Input YES", "YES\n", "Test YES", true},
		{"Input NO", "NO\n", "Test NO", false},
		{"Invalid then y", "maybe\nyes\n", "Test Invalid Y", true},
		{"Invalid then n", "perhaps\nno\n", "Test Invalid N", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			// The following lines are for capturing stdout if needed, not strictly necessary
			// if only testing return value.
			// oldStdout := os.Stdout
			// tempStdoutFile, _ := ioutil.TempFile("", "stdout")
			// os.Stdout = tempStdoutFile
			
			os.Stdin = r

			_, err = w.WriteString(tt.input)
			if err != nil {
				t.Fatalf("Failed to write to pipe: %v", err)
			}
			w.Close()

			if got := duckerlib.AsksAreYouSure(tt.prompt); got != tt.expected { // Prefixed
				t.Errorf("duckerlib.AsksAreYouSure(%q) with input %q = %v, want %v", tt.prompt, strings.TrimSpace(tt.input), got, tt.expected)
			}
			
			os.Stdin = originalStdin // Restore for next sub-test
			r.Close()
			// tempStdoutFile.Close()
			// os.Stdout = oldStdout // Restore stdout
			// os.Remove(tempStdoutFile.Name()) // Clean up temp stdout file
		})
	}
}

func TestCheckError(t *testing.T) {
	// Test with nil error (should not panic or exit)
	// As before, testing the os.Exit path is complex for a unit test.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("duckerlib.CheckError(nil) panicked: %v", r)
		}
	}()
	duckerlib.CheckError(nil) // Prefixed

	t.Log("duckerlib.CheckError(nil) passed. Testing the os.Exit path is skipped.")
}
