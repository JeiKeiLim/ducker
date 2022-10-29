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

func runTerminalCmd(cmd string, option string) string {
	cmdRun, cmdOut := exec.Command(cmd, strings.Split(option, " ")...), new(strings.Builder)
	cmdRun.Stdout = cmdOut
	err := cmdRun.Run()
    
    checkError(err)
	return strings.TrimSpace(cmdOut.String())
}

func runCommandInShell(cmd string) {
    cmdResult := exec.Command("/bin/sh", "-c", cmd)
    cmdResult.Stdout = os.Stdout
    cmdResult.Stderr = os.Stderr
    cmdResult.Stdin = os.Stdin
	if err := cmdResult.Run(); err != nil {
		fmt.Println(err)
	}
}

func getArchType() string {
	return runTerminalCmd("uname", "-m")
}

func writeFile(contents string, path string, overwrite bool) bool {
	_, err := os.Stat(path)
	if !overwrite && err == nil {
		fmt.Printf("%s already exist!\n", path)
		return false
	}

	fp, err := os.Create(path)
	checkError(err)
	write_buf := bufio.NewWriter(fp)
	_, err = write_buf.WriteString(contents)
	write_buf.Flush()

	checkError(err)

	return true
}

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

func asksAreYouSure(msg string) bool {
    reader := bufio.NewReader(os.Stdin)

    for true {
        fmt.Printf("%s (y/n) ", msg)
        keyIn, err := reader.ReadString('\n')
        checkError(err)
        keyIn = strings.TrimSpace(keyIn)
        keyIn = strings.ToLower(keyIn)
        if  keyIn == "y" || keyIn == "yes" {
            return true
        } else if keyIn == "n" || keyIn == "no" {
            return false
        }
    }

    return false
}
