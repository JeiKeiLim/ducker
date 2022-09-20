package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "net/http"
    "log"
)

func getContentFromURL(url string) string {
    resp, err := http.Get(url)

    if err != nil {
        return ""
    } else {
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return ""
        } else {
            return string(body)
        }
    }
}

func checkTemplates(template string) string {
    defaultTemplate := make(map[string]string)
    defaultTemplate["python"] = "https://raw.githubusercontent.com/JeiKeiLim/hocker/f0934622f9d75449d1b0abc2767863999f7c386a/docker/Dockerfile"
    defaultTemplate["cpp"] = "https://raw.githubusercontent.com/JeiKeiLim/hocker/f0934622f9d75449d1b0abc2767863999f7c386a/docker/Dockerfile"

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
    } else {
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
            fmt.Print("Template successfully read from ")
            log.Println(defaultURL)
        }
    }

    return ""
}
