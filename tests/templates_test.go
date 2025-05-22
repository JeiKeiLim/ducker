package duckerlib_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jeikeilim/ducker/internal/duckerlib"
)

// Note: The local `var mockGetContentFromURL` has been removed.
// Tests will now modify `duckerlib.GetContentFromURLFunc`.

func TestCheckTemplates_EmptyInput(t *testing.T) {
	originalFunc := duckerlib.GetContentFromURLFunc
	defer func() { duckerlib.GetContentFromURLFunc = originalFunc }()

	duckerlib.GetContentFromURLFunc = func(s string) string {
		t.Errorf("GetContentFromURLFunc was called for empty input with URL: %s", s)
		return "unexpected call"
	}

	if result := duckerlib.CheckTemplates(""); result != "" { // Prefixed
		t.Errorf("duckerlib.CheckTemplates(\"\") = %q, want \"\"", result)
	}
}

func TestCheckTemplates_LocalFile(t *testing.T) {
	content := "This is a test file content."
	tmpfile, err := ioutil.TempFile("", "template_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		tmpfile.Close()
		t.Fatal(err)
	}
	tmpfile.Close()

	// For local file, GetContentFromURLFunc should not be called.
	originalFunc := duckerlib.GetContentFromURLFunc
	defer func() { duckerlib.GetContentFromURLFunc = originalFunc }()
	duckerlib.GetContentFromURLFunc = func(s string) string {
		t.Errorf("GetContentFromURLFunc was called for local file with URL: %s", s)
		return "unexpected call"
	}


	if result := duckerlib.CheckTemplates(tmpfile.Name()); result != content { // Prefixed
		t.Errorf("duckerlib.CheckTemplates(localfile) = %q, want %q", result, content)
	}
}

func TestCheckTemplates_URL(t *testing.T) {
	serverContent := "Hello from mock server!"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, serverContent)
	}))
	defer server.Close()

	originalFunc := duckerlib.GetContentFromURLFunc
	defer func() { duckerlib.GetContentFromURLFunc = originalFunc }()

	// Test with a valid URL
	duckerlib.GetContentFromURLFunc = func(url string) string {
		if url == server.URL {
			resp, err := http.Get(url)
			if err != nil {
				t.Logf("Mock Get (valid URL) failed: %v", err)
				return ""
			}
			defer resp.Body.Close()
			body, errRead := ioutil.ReadAll(resp.Body)
			if errRead != nil {
				t.Logf("Mock ReadAll (valid URL) failed: %v", errRead)
				return ""
			}
			return strings.TrimSpace(string(body))
		}
		t.Errorf("Mock GetContentFromURLFunc called with unexpected URL: %s", url)
		return "mock error: unexpected url"
	}

	if result := duckerlib.CheckTemplates(server.URL); result != serverContent { // Prefixed
		t.Errorf("duckerlib.CheckTemplates(valid_url) = %q, want %q", result, serverContent)
	}

	// Test with an invalid/unreachable URL
	duckerlib.GetContentFromURLFunc = func(url string) string {
		if url == "http://invalid-url-that-will-fail" {
			return "" // Simulate error or empty response
		}
		t.Errorf("Mock GetContentFromURLFunc (for invalid) called with unexpected URL: %s", url)
		return "mock error: unexpected url for invalid test"
	}
	if result := duckerlib.CheckTemplates("http://invalid-url-that-will-fail"); result != "" { // Prefixed
		t.Errorf("duckerlib.CheckTemplates(invalid_url) = %q, want \"\"", result)
	}
}

func TestCheckTemplates_DefaultTemplate(t *testing.T) {
	pythonTemplateContent := "FROM python:3.9-slim\nWORKDIR /app\nCOPY . ."
	pythonTemplateKey := "python"
	expectedPythonURL := "https://raw.githubusercontent.com/JeiKeiLim/ducker/3076806435752ff5a0c3458ccb9ebc12553c44ea/templates/python.Dockerfile"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, pythonTemplateContent)
	}))
	defer mockServer.Close()

	originalFunc := duckerlib.GetContentFromURLFunc
	defer func() { duckerlib.GetContentFromURLFunc = originalFunc }()

	duckerlib.GetContentFromURLFunc = func(url string) string {
		if url == expectedPythonURL {
			resp, err := http.Get(mockServer.URL)
			if err != nil {
				t.Fatalf("Mock http.Get to mockServer failed: %v", err)
				return ""
			}
			defer resp.Body.Close()
			body, errRead := ioutil.ReadAll(resp.Body)
			if errRead != nil {
				t.Fatalf("Mock ioutil.ReadAll from mockServer failed: %v", errRead)
				return ""
			}
			return strings.TrimSpace(string(body))
		}
		t.Errorf("Mock GetContentFromURLFunc (for default) called with unexpected URL: %s, expected %s", url, expectedPythonURL)
		return ""
	}
	
	if result := duckerlib.CheckTemplates(pythonTemplateKey); result != pythonTemplateContent { // Prefixed
		t.Errorf("duckerlib.CheckTemplates(%q) = %q, want %q", pythonTemplateKey, result, pythonTemplateContent)
	}
	
	// Test with a non-existent default template
	duckerlib.GetContentFromURLFunc = func(url string) string {
		t.Errorf("GetContentFromURLFunc was called for non-existent default template key with URL: %s", url)
		return "unexpected call for non-existent default"
	}
	if result := duckerlib.CheckTemplates("nonexistenttemplatekey"); result != "" { // Prefixed
		t.Errorf("duckerlib.CheckTemplates(\"nonexistenttemplatekey\") = %q, want \"\"", result)
	}
}


func TestCheckTemplates_NonExistent(t *testing.T) {
	originalFunc := duckerlib.GetContentFromURLFunc
	defer func() { duckerlib.GetContentFromURLFunc = originalFunc }()
	
	duckerlib.GetContentFromURLFunc = func(url string) string {
		t.Errorf("GetContentFromURLFunc was called unexpectedly for non-existent input with URL: %s", url)
		return "unexpected call"
	}

	input := "this-is-not-a-file-url-or-template"
	if result := duckerlib.CheckTemplates(input); result != "" { // Prefixed
		t.Errorf("duckerlib.CheckTemplates(%q) = %q, want \"\"", input, result)
	}
}
