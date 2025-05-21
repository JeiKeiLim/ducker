package duckerlib

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// GetContentFromURLFunc is a function variable that CheckTemplates uses to fetch URL content.
// It defaults to the actual GetContentFromURL but can be replaced by tests for mocking.
var GetContentFromURLFunc func(url string) string = GetContentFromURL

// CheckTemplates checks if a template is a local file, a URL, or a default template.
func CheckTemplates(template string) string {
	defaultTemplate := make(map[string]string)
	defaultTemplate["python"] = "https://raw.githubusercontent.com/JeiKeiLim/ducker/3076806435752ff5a0c3458ccb9ebc12553c44ea/templates/python.Dockerfile"
	defaultTemplate["cpp"] = "https://raw.githubusercontent.com/JeiKeiLim/ducker/3076806435752ff5a0c3458ccb9ebc12553c44ea/templates/cpp.Dockerfile"

	if template == "" {
		return ""
	}

	if _, err := os.Stat(template); err == nil {
		fmt.Print("Template exist! Using ")
		fmt.Println(template)

		content, err := ioutil.ReadFile(template)

		if err != nil {
			log.Fatal(err)
		}

		return string(content)
	}
	// Check if template is a URL or a default template key
	var urlToFetch string
	var isDefaultKey bool = false // To log if we are using a default template URL

	if strings.HasPrefix(template, "http://") || strings.HasPrefix(template, "https://") {
		urlToFetch = template
	} else if defaultURLValue, ok := defaultTemplate[strings.ToLower(template)]; ok { // Use ToLower for case-insensitive key match
		urlToFetch = defaultURLValue
		isDefaultKey = true
	}

	if urlToFetch != "" {
		if isDefaultKey {
			// Log that we are using a default template URL, consistent with original log messages
			log.Println("Using default template URL:", urlToFetch)
		}
		// Use the GetContentFromURLFunc function variable
		urlContent := GetContentFromURLFunc(urlToFetch)
		if urlContent != "" {
			fmt.Print("Template successfully read from ")
			fmt.Println(urlToFetch) // Log the actual URL fetched

			return urlContent
		}
	}

	return ""
}
