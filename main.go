package main

import (
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

	if err := proc.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		} else {
			panic(err)
		}
	}
}

func main() {
	aproxy := os.Getenv("APROXY")
	if len(aproxy) == 0 {
		runCMD(os.Args[1], os.Args[2:])

		return
	}

	os.Setenv("ALL_PROXY", aproxy)
	os.Setenv("http_proxy", aproxy)
	os.Setenv("https_proxy", aproxy)
	os.Setenv("ftp_proxy", aproxy)
	os.Setenv("GOPROXY", "https://goproxy.cn")

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

	runCMD(os.Args[1], os.Args[1:])
}
