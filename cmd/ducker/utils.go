package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Run terminal command and return its output.
func getTerminalCmdOut(cmd string, option string) string {
	cmdRun, cmdOut := exec.Command(cmd, strings.Split(option, " ")...), new(strings.Builder)
	cmdRun.Stdout = cmdOut
	err := cmdRun.Run()

	checkError(err)
	return strings.TrimSpace(cmdOut.String())
}

// Run terminall command in shell with user interaction enabled.
func runTerminalCmdInShell(cmd string) {
	cmdResult := exec.Command("/bin/sh", "-c", cmd)
	cmdResult.Stdout = os.Stdout
	cmdResult.Stderr = os.Stderr
	cmdResult.Stdin = os.Stdin
	if err := cmdResult.Run(); err != nil {
		fmt.Println(err)
	}
}

func getArchType() string {
	return getTerminalCmdOut("uname", "-m")
}

func writeFile(contents string, path string, overwrite bool) bool {
	_, err := os.Stat(path)
	if !overwrite && err == nil {
		fmt.Printf("%s already exist!\n", path)
		return false
	}

	fp, err := os.Create(path)
	checkError(err)
	writeBuf := bufio.NewWriter(fp)
	_, err = writeBuf.WriteString(contents)
	writeBuf.Flush()

	checkError(err)

	return true
}

func getContentFromURL(url string) string {
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

func asksAreYouSure(msg string) bool {
	reader := bufio.NewReader(os.Stdin)

	for true {
		fmt.Printf("%s (y/n) ", msg)
		keyIn, err := reader.ReadString('\n')
		checkError(err)
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
