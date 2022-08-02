// Please refer to https://cli.urfave.org/v2

package main

import (
	"bufio"
	"fmt"
	"log"
    "time"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thlib/go-timezone-local/tzlocal"
	"github.com/urfave/cli/v2"
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
	cmdRun.Run()

	return strings.TrimSpace(cmdOut.String())
}

func getArchType() string {
	return runTerminalCmd("uname", "-m")
}

func writeFile(contents string, path string) {
	fp, err := os.Create(path)
	checkError(err)
	write_buf := bufio.NewWriter(fp)
	write_buf.WriteString(contents)
	write_buf.Flush()
}

func dockerBuild(ctx *cli.Context, dockerTag string) {
	dockerFilePath := "docker/Dockerfile"
	if !strings.HasSuffix(dockerTag, "x86_64") {
		dockerFilePath += "." + getArchType()
	}

	buildCmd := "docker build . -t " + dockerTag
	buildCmd += " -f " + dockerFilePath
	buildCmd += " --build-arg UID=" + runTerminalCmd("id", "-u")
	buildCmd += " --build-arg GID=" + runTerminalCmd("id", "-g")

	cmdRun := exec.Command("/bin/sh", "-c", buildCmd)
	cmdRun.Stdout = os.Stdout
	if err := cmdRun.Run(); err != nil {
		fmt.Println(err)
	}
}

func dockerRun(ctx *cli.Context, dockerTag string) {
	dockerArgs := ""
	shellCmd := "/bin/bash"

	runCmd := "docker run -tid --privileged"
	runCmd += " -e DISPLAY=" + os.Getenv("DISPLAY")
	runCmd += " -e TERM=xterm-256color"
	runCmd += " -v /tmp/.X11-unix:/tmp/.X11-unix:ro"
	runCmd += " -v /dev:/dev"

    // Mount if .gitconfig exists
    gitConfigPath := filepath.Join(os.Getenv("HOME"), ".gitconfig")
    if _, err := os.Stat(gitConfigPath); err == nil {
        runCmd += " -v " + gitConfigPath + ":/home/user/.gitconfig"
    }

	runCmd += " --network host"
	runCmd += " " + dockerArgs
	runCmd += " " + dockerTag
	runCmd += " " + shellCmd

	cmdRun := exec.Command("/bin/sh", "-c", runCmd)
	cmdRun.Stdout = os.Stdout
	cmdRun.Stderr = os.Stderr
	cmdRun.Stdin = os.Stdin
	if err := cmdRun.Run(); err != nil {
		fmt.Println(err)
	}

	lastContainerID := runTerminalCmd("docker", "ps -qn 1")
	writeFile(lastContainerID, ".last_exec_cont_id.txt")

	dockerExec()
}

func dockerExec() {
	lastContainerID := runTerminalCmd("tail", "-1 .last_exec_cont_id.txt")
	shellCmd := "/bin/bash"

	execCmd := "docker exec -ti " + lastContainerID
	execCmd += " " + shellCmd

	cmdExec := exec.Command("/bin/sh", "-c", execCmd)
	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr
	cmdExec.Stdin = os.Stdin
	if err := cmdExec.Run(); err != nil {
		fmt.Println(err)
	}

}

func initDockerfile(ctx *cli.Context) {
	name := ctx.String("name")
	contact := ctx.String("contact")

	if !ctx.Bool("quite") {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter your name(Default: %s): ", name)
		keyIn, err := reader.ReadString('\n')
		checkError(err)
		if strings.TrimSpace(keyIn) != "" {
			name = keyIn
		}
		name = strings.TrimSpace(name)

		fmt.Printf("Enter contact address (Default: %s): ", contact)
		keyIn, err = reader.ReadString('\n')
		checkError(err)
		if strings.TrimSpace(keyIn) != "" {
			contact = keyIn
		}
		contact = strings.TrimSpace(contact)
	}

	tzname, _ := tzlocal.RuntimeTZ()

	dockerBaseImage := "ubuntu:bionic"

	dockerContents := fmt.Sprintf("FROM %s\n\n", dockerBaseImage)

	dockerContents += fmt.Sprintf("LABEL maintainer=\"%s <%s>\"\n", name, contact)
	dockerContents += fmt.Sprintf("ENV DEBIAN_FRONTEND=noninteractive\n")
	dockerContents += fmt.Sprintf("ENV TZ=%s\n", tzname)
	dockerContents += "ENV TERM xterm-256color\n\n"

	dockerContents += "ARG UID=1000\n"
	dockerContents += "ARG GID=1000\n"
	dockerContents += "RUN groupadd -g $GID -o user && useradd -m -u $UID -g $GID -o -s /bin/bash user\n\n"

	dockerContents += "RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone\n"
	dockerContents += "RUN apt-get update && apt-get install -y sudo dialog apt-utils tzdata\n"
	dockerContents += "RUN echo \"%sudo ALL=(ALL) NOPASSWD: ALL\" >> /etc/sudoers && echo \"user:user\" | chpasswd && adduser user sudo\n\n"

	dockerContents += "WORKDIR /home/user\n"
	dockerContents += "USER user\n\n"

	dockerContents += "ENV NVIDIA_VISIBLE_DEVICES ${NVIDIA_VISIBLE_DEVICES:-all}\n"
	dockerContents += "ENV NVIDIA_DRIVER_CAPABILITIES ${NVIDIA_DRIVER_CAPABILITIES:+$NVIDIA_DRIVER_CAPABILITIES,}graphics\n\n"

    dockerContents += "RUN sudo apt-get update && sudo apt-get -y install wget curl git\n"
    dockerContents += "RUN curl -s https://raw.githubusercontent.com/JeiKeiLim/my_term/main/run.sh | /bin/bash\n\n"

	dockerContents += "# Place your environment here\n\n"

	os.Mkdir("docker", os.ModePerm)

	writeFile(dockerContents, "docker/Dockerfile")
	writeFile(dockerContents, "docker/Dockerfile.aarch64")

	fmt.Println("Success!")
	fmt.Println("Dockerfile has been created in docker directory")
}

func main() {
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	organizationName := "jeikeilim"
	baseDir := filepath.Base(mydir)
	projectName := strings.ToLower(baseDir)

	osArchType := getArchType()
	dockerTag := fmt.Sprintf("%s/%s:%s", organizationName, projectName, osArchType)

	app := &cli.App{
		Name:  "hocker",
        Version: "0.1.0",
        Compiled: time.Now(),
        Authors: []*cli.Author{
            &cli.Author{
                Name: "Jongkuk Lim",
                Email: "lim.jeikei@gmail.com",
            },
        },
        EnableBashCompletion: true,
		Usage: "Hocker the docker helper",
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Init Dockerfile setting in this directory",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "Your name",
						Value:       "Anonymous",
						DefaultText: "Anonymous",
					},
					&cli.StringFlag{
						Name:        "contact",
						Aliases:     []string{"c"},
						Usage:       "Contact address",
						Value:       "None",
						DefaultText: "None",
					},
					&cli.BoolFlag{
						Name:    "quite",
						Aliases: []string{"q"},
						Usage:   "Do not use interactive mode",
					},
				},
				Action: func(cCtx *cli.Context) error {
					initDockerfile(cCtx)
					return nil
				},
			},
			{
				Name:    "build",
				Aliases: []string{"b"},
				Usage:   "Build docker image",
				Action: func(cCtx *cli.Context) error {
					dockerBuild(cCtx, dockerTag)
					return nil
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Running docker image",
				Action: func(cCtx *cli.Context) error {
					dockerRun(cCtx, dockerTag)
					return nil
				},
			},
			{
				Name:    "exec",
				Aliases: []string{"r"},
				Usage:   "Executing the docker container",
				Action: func(cCtx *cli.Context) error {
					dockerExec()
					return nil
				},
			},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("Hello %q", cCtx.Args().Get(0))
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
