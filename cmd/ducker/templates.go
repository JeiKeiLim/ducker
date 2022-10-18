package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func checkTemplates(template string) string {
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
	// Check URL exist.
	urlContent := getContentFromURL(template)

	if urlContent != "" {
		fmt.Print("Template successfully read from ")
		fmt.Println(template)
		return urlContent
	}

	// Check if template name has default setting
	defaultURL := defaultTemplate[template]
	if defaultURL == "" {
		log.Println("Template can not be found")
	} else {
		urlContent = getContentFromURL(defaultURL)

		if urlContent != "" {
			fmt.Print("Template successfully read from ")
			log.Println(defaultURL)

			return urlContent
		}
	}

	return ""
}
