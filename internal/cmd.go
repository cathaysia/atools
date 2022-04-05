package internal

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrHOMENotExists = errors.New("HOME nev doesn't exists")
	Errshell         = errors.New("Shell cannot recognition")
)

func ErrAPROXY(msg string, err error) error {
	return fmt.Errorf("%v: %w", msg, err)
}

func GetCurrentShell() string {
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
