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
	CheckError(err)
	writeBuf := bufio.NewWriter(fp)
	_, err = writeBuf.WriteString(contents)
	writeBuf.Flush()

	CheckError(err)

	return true
}

// GetContentFromURL fetches content from a URL.
func GetContentFromURL(url string) string {
	resp, err := http.Get(url)

	if err != nil {
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
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
