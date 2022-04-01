package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runCMD(cmd string, args []string) {
	// TODO: 检查 cmd 是否是别名
	if lp, err := exec.LookPath(cmd); err != nil {
		fmt.Println(lp)
		// 如果 cmd 没办法找到
		cmd = os.Getenv("SHELL")
		args = []string{
			cmd, "-c", `'` + strings.Join(args, " ") + `'`,
		}
	} else {
		cmd = lp
		args[0] = lp
	}

	proc := exec.Command(cmd, "")
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	proc.Args = args

	err := proc.Run()
	if err == nil {
		return
	}

	var exitError exec.ExitError

	if ok := errors.As(err, &exitError); ok {
		os.Exit(exitError.ExitCode())
	}

	panic(err)
}

func checkArgs() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: aproxy [<command>]")

		os.Exit(0)
	}
}

func setEnvByMap(envs map[string]string) error {
	for name, value := range envs {
		if err := os.Setenv(name, value); err != nil {
			return err
		}
	}

	return nil
}

func setAlias(aproxy string) {
	switch os.Args[1] {
	case "curl":
		os.Args = append(os.Args, "--proxy", aproxy)
	case "git":
		os.Args = append([]string{
			os.Args[0],
			os.Args[1],
			"-c", fmt.Sprintf("http.proxy=%s", aproxy),
			"-c", fmt.Sprintf("https.proxy=%s", aproxy),
			"-c", "http.sslVerify=false",
			"-c", "https.sslVerify=false",
		}, os.Args[2:]...)
	case "svn":
		// TODO: 这个还没测试
		pos := strings.Index(aproxy, ":")
		os.Args = append([]string{
			os.Args[0],
			os.Args[1],
			"--config-option", fmt.Sprintf("servers:global:http-proxy-host=%s", aproxy[:pos]),
			"--config-option", fmt.Sprintf("servers:global:http-proxy-port=%s", aproxy[pos:]),
		}, os.Args[2:]...)
	}
}

func main() {
	checkArgs()

	aproxy, ok := os.LookupEnv("APROXY")
	if !ok {
		runCMD(os.Args[1], os.Args[2:])

		return
	}

	envs := map[string]string{
		"ALL_PROXY":   aproxy,
		"http_proxy":  aproxy,
		"https_proxy": aproxy,
		"ftp_proxy":   aproxy,
		"GOPROXY":     "https://goproxy.cn",
	}

	if err := setEnvByMap(envs); err != nil {
		panic(err)
	}

	runCMD(os.Args[1], os.Args[1:])
}
