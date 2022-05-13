package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"atools/internal"

	"github.com/sirupsen/logrus"
)

func main() {
	checkArgs()
	logrus.SetLevel(logrus.TraceLevel)

	aproxy, ok := os.LookupEnv("APROXY")
	if !ok {
		err := runCMD(os.Args[1], os.Args[2:])
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}

		if err != nil {
			panic(err)
		}

		return
	}

	envs := map[string]string{
		"ALL_PROXY":   aproxy,
		"http_proxy":  aproxy,
		"https_proxy": aproxy,
		"ftp_proxy":   aproxy,
		"GOPROXY":     "https://goproxy.cn",
	}

	if err := setEnvs(envs); err != nil {
		panic(err)
	}

	// replaceCMD(aproxy)
	os.Args = append([]string{
		os.Args[0],
		os.Args[1],
	}, append(getExtraArgsForCMD(os.Args[1], aproxy), os.Args[2:]...)...)

	err := runCMD(os.Args[1], os.Args[2:])
	if exitError, ok := err.(*exec.ExitError); ok {
		os.Exit(exitError.ExitCode())
	}

	if err != nil {
		panic(err)
	}
}

func runCMD(cmd string, args []string) error {
	// cmd 有三种情况：
	// 1. cmd 是一个可在 PATH 中找到的二进制文件
	// 2. cmd 是一个 shell 的 alias
	// 3. cmd 是一个 shell 命令
	// 后两种的处理方式相同：直接运行 zsh -c 'source ~/.zshrc && cmd'
	currentShell, err := exec.LookPath(cmd)
	if err != nil {
		currentShell = internal.GetCurrentShell()

		profile, err := internal.GetShellProfile(currentShell)
		if err != nil {
			return err
		}

		args = []string{
			"-c",
			fmt.Sprintf(`"source %v && %v %v"`, profile, cmd, strings.Join(args, " ")),
		}
	}

	cmd = currentShell

	proc := exec.Command(cmd, args...)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin

	logrus.Debugf("proc: %v", proc)

	return proc.Run()
}

func checkArgs() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: aproxy [<command>]")

		os.Exit(0)
	}
}

func setEnvs(envs map[string]string) error {
	for name, value := range envs {
		if err := os.Setenv(name, value); err != nil {
			return err
		}
	}

	return nil
}

func getExtraArgsForCMD(cmd string, aproxy string) []string {
	if cmd == "curl" {
		return []string{
			"--proxy",
			aproxy,
		}
	}

	if cmd == "git" {
		return []string{
			"-c", fmt.Sprintf("http.proxy=%s", aproxy),
			"-c", fmt.Sprintf("https.proxy=%s", aproxy),
			"-c", "http.sslVerify=false",
			"-c", "https.sslVerify=false",
		}
	}

	if cmd == "svn" {
		// TODO: 这个还没测试
		pos := strings.Index(aproxy, ":")

		return []string{
			"--config-option", fmt.Sprintf("servers:global:http-proxy-host=%s", aproxy[:pos]),
			"--config-option", fmt.Sprintf("servers:global:http-proxy-port=%s", aproxy[pos:]),
		}
	}

	return []string{}
}
