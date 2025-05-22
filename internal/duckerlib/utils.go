package duckerlib

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// CheckError prints an error message and exits if the error is not nil.
func CheckError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// GetTerminalCmdOut runs a terminal command and return its output.
func GetTerminalCmdOut(cmd string, option string) string {
	cmdRun, cmdOut := exec.Command(cmd, strings.Split(option, " ")...), new(strings.Builder)
	cmdRun.Stdout = cmdOut
	err := cmdRun.Run()

	CheckError(err)
	return strings.TrimSpace(cmdOut.String())
}

// RunTerminalCmdInShell runs a terminal command in shell with user interaction enabled.
func RunTerminalCmdInShell(cmd string) {
	cmdResult := exec.Command("/bin/sh", "-c", cmd)
	cmdResult.Stdout = os.Stdout
	cmdResult.Stderr = os.Stderr
	cmdResult.Stdin = os.Stdin
	if err := cmdResult.Run(); err != nil {
		fmt.Println(err)
	}
}

// GetArchType returns the system's architecture type using "uname -m".
func GetArchType() string {
	return GetTerminalCmdOut("uname", "-m")
}

// WriteFile writes contents to a file at the given path.
// It handles overwriting based on the overwrite flag.
func WriteFile(contents string, path string, overwrite bool) bool {
	_, err := os.Stat(path)
	if !overwrite && err == nil {
		fmt.Printf("%s already exist!\n", path)
		return false
	}

	fp, err := os.Create(path)
	if err != nil { // Added error check for os.Create
		CheckError(err) // Or handle more gracefully
		return false    // Return false if create fails
	}
	defer fp.Close() // Ensure file is closed

	writeBuf := bufio.NewWriter(fp)
	_, err = writeBuf.WriteString(contents)
	if err != nil { // Added error check
		CheckError(err) // Or handle
		return false
	}
	err = writeBuf.Flush()
	if err != nil { // Added error check
		CheckError(err) // Or handle
		return false
	}
	return true
}

// GetContentFromURL fetches content from a URL.
func GetContentFromURL(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return "" // Error during GET request itself
	}
	defer resp.Body.Close() // Ensure body is always closed

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Not a successful status code (not 2xx)
		return "" // Return empty string as per test expectation
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// Error reading the body, even if status was success
		return ""
	}

	return string(body)
}

// AsksAreYouSure prompts the user with a yes/no question and returns their choice.
func AsksAreYouSure(msg string) bool {
	reader := bufio.NewReader(os.Stdin)

	for true {
		fmt.Printf("%s (y/n) ", msg)
		keyIn, err := reader.ReadString('\n')
		CheckError(err)
		keyIn = strings.TrimSpace(keyIn)
		keyIn = strings.ToLower(keyIn)
		if keyIn == "y" || keyIn == "yes" {
			return true
		} else if keyIn == "n" || keyIn == "no" {
			return false
		}
	}

	return false
}
