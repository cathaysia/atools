package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	ErrHOMENotExists = errors.New("HOME nev doesn't exists")
	Errshell         = errors.New("Shell cannot recognition")
	ErrSystem        = errors.New("System Not Support")
)

func ErrAPROXY(msg string, err error) error {
	return fmt.Errorf("%v: %w", msg, err)
}

func GetProcessNameByPID(pid uint64) (string, error) {
	if _, err := os.Stat("/proc"); err != nil {
		return "", ErrSystem
	}

	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%v/comm", pid))
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(string(data), "\n", ""), nil
}

func GetCurrentShell() string {
	parentName, err := GetProcessNameByPID(uint64(os.Getppid()))
	if err == nil {
		logrus.Debugf("The parent process of aproxy is %v", parentName)

		if strings.Contains(parentName, "zsh") || strings.Contains(parentName, "bash") || strings.Contains(parentName, "fish") {
			if path, err := exec.LookPath(parentName); err == nil {
				return path
			}
		}
	}

	if shell, ok := os.LookupEnv("SHELL"); ok {
		return shell
	}

	return "/bin/sh"
}

func GetShellProfile(shell string) (string, error) {
	home, ok := os.LookupEnv("HOME")

	if !ok {
		return "", ErrHOMENotExists
	}

	if strings.Contains(shell, "bash") {
		return fmt.Sprintf("%v/.bashrc", home), nil
	}

	if strings.Contains(shell, "zsh") {
		return fmt.Sprintf("%v/.zshrc", home), nil
	}

	return "", Errshell
}
