// Please refer to https://cli.urfave.org/v2
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thlib/go-timezone-local/tzlocal"
	"github.com/urfave/cli/v2"
)

var (
    version string
)

// Cross compile options
// env GOOS=darwin GOARCH=arm64 go build ./cmd/ducker

func duckerConfig(ctx *cli.Context) {
	// TODO(jeikeilim): add more flexible setting options
	genGlobal := ctx.Bool("global")
	genLocal := ctx.Bool("local")

	if genGlobal {
		doWrite := true
		defaultGlobalConfig := getDefaultGlobalConfig("", "", "")

		if _, err := os.Stat(getDefaultGlobalConfigPath()); err == nil {
			doWrite = asksAreYouSure("Ducker global config file already exists at " + getDefaultGlobalConfigPath() + " Do you want to OVERWRITE?")
		}
		if doWrite {
			defaultGlobalConfig.Write(getDefaultGlobalConfigPath())
			fmt.Println("Global config has been written in", getDefaultGlobalConfigPath())
		}
	}

	if genLocal {
		doWrite := true
		defaultLocalConfig := getDefaultLocalConfig()
		if _, err := os.Stat(getDefaultLocalConfigPath()); err == nil {
			doWrite = asksAreYouSure("Ducker local config file already exists at " + getDefaultLocalConfigPath() + " Do you want to OVERWRITE?")
		}

		if doWrite {
			defaultLocalConfig.Write(getDefaultLocalConfigPath())
			fmt.Println("Local config has been written in", getDefaultLocalConfigPath())
		}
	}
}

func dockerBuild(ctx *cli.Context, dockerTag string) {
	dockerFilePath := "docker/Dockerfile"
	if !strings.HasSuffix(dockerTag, "x86_64") {
		dockerFilePath += "." + getArchType()
	}
	localConfig := readDefaultLocalConfig()

	buildArgs := ctx.String("args")

	buildCmd := "docker build . -t " + dockerTag
	buildCmd += " -f " + dockerFilePath
	buildCmd += " " + localConfig.GetBuildArg()

	if buildArgs != "" {
		buildCmd += " " + buildArgs
	}

	fmt.Println(buildCmd)

	runTerminalCmdInShell(buildCmd)
}

func dockerRun(ctx *cli.Context, dockerTag string) {
	dockerArgs := ctx.String("docker-args")
	shellType := ctx.String("shell")
	mountPWD := ctx.Bool("mount-pwd")
	shellCmd := "/bin/bash"
	dockerOpt := "-tid"

	runOption := ""
	localConfig := readDefaultLocalConfig()
	if !localConfig.IsEmpty() {
		runOption += localConfig.GetRunArg()
	}

	if shellType == "zsh" {
		shellCmd = "/usr/bin/zsh"
	} else if shellType == "nosh" {
		shellCmd = ctx.String("run-cmd")
		dockerOpt = "-ti"
	}

	runCmd := "docker run " + dockerOpt + " "
	runCmd += runOption

	if mountPWD || localConfig.Mount_PWD {
		mydir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		} else {
			baseDir := filepath.Base(mydir)
			projectName := strings.ToLower(baseDir)

			runCmd += " -v " + mydir + ":/home/user/" + projectName
		}
	}

	// Add GPU option if NVIDIA driver exist
	if _, err := os.Stat("/proc/driver/nvidia/version"); err == nil {
		runCmd += " --gpus all"
	}

	// Mount if .gitconfig exists
	gitConfigPath := filepath.Join(os.Getenv("HOME"), ".gitconfig")
	if _, err := os.Stat(gitConfigPath); err == nil {
		runCmd += " -v " + gitConfigPath + ":/home/user/.gitconfig"
	}

	runCmd += " " + dockerArgs
	runCmd += " " + dockerTag
	runCmd += " " + shellCmd

	fmt.Println(runCmd)

	runTerminalCmdInShell(runCmd)

	// TODO(jeikeilim): It's bad idea to check last container ID with docker ps -qn 1
	lastContainerID := getTerminalCmdOut("docker", "ps -qn 1")
	localConfig.LastExecID = lastContainerID
	localConfig.Write(getDefaultLocalConfigPath())

	if shellType != "nosh" {
		dockerExec(ctx)
	}
}

func dockerExec(ctx *cli.Context) {
	localConfig := readDefaultLocalConfig()

	shellType := ctx.String("shell")
	shellCmd := "/bin/bash"

	if shellType == "zsh" {
		shellCmd = "/usr/bin/zsh"
	}

	lastContainerID := localConfig.LastExecID

	if lastContainerID == "" {
		fmt.Println("Last container ID can not be found.")
		return
	}

	execCmd := "docker exec -ti " + lastContainerID
	execCmd += " " + shellCmd

	runTerminalCmdInShell(execCmd)
}

func dockerPs() {
	// TODO(jeikeilim): add option to show $PWD containers only.
	runTerminalCmdInShell("docker ps")
}

func dockerLs() {
	runTerminalCmdInShell("docker images")
}

func duckerStop(ctx *cli.Context, dockerTag string) {
	psString := "ps -f ancestor=" + dockerTag + " -q"
	outContainerID := getTerminalCmdOut("docker", psString)
	list_outContainerID := strings.Split(outContainerID, "\n")

	// TODO(ulken94): Add option to choose which container to kill.
	cnt := 0
	for _, cID := range list_outContainerID {
		if strings.TrimSpace(cID) == "" {
			break
		}
		cnt += 1
		killString := "kill " + cID
		getTerminalCmdOut("docker", killString)
	}
	fmt.Printf("Total %d container(s) is(are) killed.\n", cnt)
}

func initDockerfile(ctx *cli.Context) {
	organization := ctx.String("organization")
	name := ctx.String("name")
	contact := ctx.String("contact")
	template := ctx.String("template")
	forceOverwrite := ctx.Bool("force")

	fmt.Println(template)
	templateContent := checkTemplates(template)

	globalConfig := readDefaultGlobalConfig()

	if !ctx.Bool("quite") && globalConfig.IsEmpty() {
		globalConfig = getDefaultGlobalConfig("", "", "")
	} else if ctx.Bool("quite") && globalConfig.IsEmpty() {
		globalConfig = getDefaultGlobalConfig(organization, name, contact)
	}

	tzname, _ := tzlocal.RuntimeTZ()

	// TODO(jeikeilim): Add custom base image support
	dockerBaseImage := "ubuntu:bionic"
	dockerContents := fmt.Sprintf("FROM %s\n\n", dockerBaseImage)

	dockerContents += fmt.Sprintf("LABEL maintainer=\"%s <%s>\"\n\n",
		globalConfig.Name,
		globalConfig.Contact)

	// TODO(jeikeilim): Modify below to use base Dockerfile
	dockerContents += "ENV DEBIAN_FRONTEND=noninteractive\n"
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

	dockerContents += "RUN sudo apt-get update && sudo apt-get install -y libgl1-mesa-dev && sudo apt-get -y install jq\n\n"

	dockerContents += "ENV NVIDIA_VISIBLE_DEVICES ${NVIDIA_VISIBLE_DEVICES:-all}\n"
	dockerContents += "ENV NVIDIA_DRIVER_CAPABILITIES ${NVIDIA_DRIVER_CAPABILITIES:+$NVIDIA_DRIVER_CAPABILITIES,}graphics\n\n"

	dockerContents += "RUN sudo apt-get update && sudo apt-get -y install wget curl git\n"
	dockerContents += "RUN curl -s https://raw.githubusercontent.com/JeiKeiLim/my_term/main/run.sh | /bin/bash\n\n"

	dockerContents += "RUN sudo apt-get update && sudo apt-get install -y zsh && \\\n"
	dockerContents += "    sh -c \"$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)\" \"\" --unattended && \\\n"
	dockerContents += "    git clone --depth=1 https://github.com/romkatv/powerlevel10k.git ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k\n"
	dockerContents += "RUN echo \"\\n# Custom settings\" >> /home/user/.zshrc && \\\n"
	dockerContents += "    echo \"export PATH=/home/user/.local/bin:$PATH\" >> /home/user/.zshrc && \\\n"
	dockerContents += "    echo \"export LC_ALL=C.UTF-8 && export LANG=C.UTF-8\" >> /home/user/.zshrc && \\\n"
	dockerContents += "    sed '11 c\\ZSH_THEME=powerlevel10k/powerlevel10k' ~/.zshrc  > tmp.txt && mv tmp.txt ~/.zshrc && \\\n"
	dockerContents += "    echo 'POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD=true' >> ~/.zshrc\n"

    // Change default shell to zsh
    dockerContents += "RUN sudo chsh -s $(which zsh) $(whoami)"

	if templateContent != "" {
		dockerContents += "\n"
		dockerContents += templateContent
	}
	dockerContents += "\n# Place your environment here\n\n"

	err := os.Mkdir("docker", os.ModePerm)
	if !forceOverwrite {
		checkError(err)
	}

	isSuccess1 := writeFile(dockerContents, "docker/Dockerfile", forceOverwrite)
	isSuccess2 := writeFile(dockerContents, "docker/Dockerfile.aarch64", forceOverwrite)

	if isSuccess1 && isSuccess2 {
		fmt.Println("Success!")
		fmt.Println("Dockerfile has been created in docker directory")
	} else {
		fmt.Println("Failed")
		fmt.Println("Dockerfile already exist in docker directory")
	}

	globalConfig.Write(getDefaultGlobalConfigPath())
	writeDefaultLocalConfig()
}

func main() {
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	globalConfig := readDefaultGlobalConfig()
	organizationName := "jeikeilim"

	if !globalConfig.IsEmpty() {
		organizationName = globalConfig.Organization
	}

	baseDir := filepath.Base(mydir)
	projectName := strings.ToLower(baseDir)

	osArchType := getArchType()
	dockerTag := fmt.Sprintf("%s/%s:%s", organizationName, projectName, osArchType)
	const duckIcon = `

=====================================
     __           __           __  
 ___( o)>     ___( o)>     ___( o)>
 \ <_. )      \ <_. )      \ <_. ) 
  '---'        '---'        '---'  
=====================================
    `

	app := &cli.App{
		Name:     "ducker",
		Version:  version,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Jongkuk Lim",
				Email: "lim.jeikei@gmail.com",
			},
		},
		EnableBashCompletion: true,
		Usage:                "Ducker the docker helper" + duckIcon,
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
					&cli.StringFlag{
						Name:        "organization",
						Aliases:     []string{"o"},
						Usage:       "Organization name",
						Value:       "None",
						DefaultText: "None",
					},
					&cli.BoolFlag{
						Name:    "quite",
						Aliases: []string{"q"},
						Usage:   "Do not use interactive mode",
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force to create Dockerfile (WARNING: file can be overwritten)",
					},
					&cli.StringFlag{
						Name:        "template",
						Aliases:     []string{"t"},
						Usage:       "Docker template path (URL, File Path, Default Templates[python, cpp])",
						Value:       "",
						DefaultText: "",
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "args",
						Aliases:     []string{"a"},
						Usage:       "Extra arguments for docker build. ex) ducker build --args \"--build-arg TEST=true\"",
						Value:       "",
						DefaultText: "",
					},
				},
				Action: func(cCtx *cli.Context) error {
					dockerBuild(cCtx, dockerTag)
					return nil
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Running docker image",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "docker-args",
						Aliases:     []string{"da"},
						Usage:       "Extra arguments for docker run. ex) ducker run --docker-args \"-v $PWD:/home/user/ducker\"",
						Value:       "",
						DefaultText: "",
					},
					&cli.StringFlag{
						Name:        "shell",
						Aliases:     []string{"s"},
						Usage:       "Shell type to run (bash, zsh, nosh)",
						Value:       "zsh",
						DefaultText: "zsh",
					},
					&cli.BoolFlag{
						Name:    "mount-pwd",
						Aliases: []string{"m"},
						Usage:   "Mount current directory to the container",
					},
					&cli.StringFlag{
						Name:        "run-cmd",
						Aliases:     []string{"r"},
						Usage:       "Running command (only applies when shell=nosh)",
						Value:       "",
						DefaultText: "",
					},
				},
				Action: func(cCtx *cli.Context) error {
					dockerRun(cCtx, dockerTag)
					return nil
				},
			},
			{
				Name:    "exec",
				Aliases: []string{"r"},
				Usage:   "Executing the docker container",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "shell",
						Aliases:     []string{"s"},
						Usage:       "Shell type to run (bash, zsh)",
						Value:       "zsh",
						DefaultText: "zsh",
					},
				},
				Action: func(cCtx *cli.Context) error {
					dockerExec(cCtx)
					return nil
				},
			},
			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Ducker config file generation",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "global",
						Aliases: []string{"g"},
						Usage:   "Generate global config file at $HOME/.ducker_global.yaml",
					},
					&cli.BoolFlag{
						Name:    "local",
						Aliases: []string{"l"},
						Usage:   "Generate local config file at $PWD/.ducker.yaml",
					},
				},
				Action: func(cCtx *cli.Context) error {
					duckerConfig(cCtx)
					return nil
				},
			},
			{
				Name:  "ps",
				Usage: "Check docker container that is running now",
				Action: func(cCtx *cli.Context) error {
					dockerPs()
					return nil
				},
			},
			{
				Name:    "images",
				Aliases: []string{"ls"},
				Usage:   "Check existing docker images",
				Action: func(cCtx *cli.Context) error {
					dockerLs()
					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Check existing docker images",
				Action: func(cCtx *cli.Context) error {
					duckerStop(cCtx, dockerTag)
					return nil
				},
			},
		},
		Action: func(cCtx *cli.Context) error {
			cli.ShowAppHelp(cCtx)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
