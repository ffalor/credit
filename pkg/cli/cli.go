package cli

import (
	"github.com/ffalor/credit/pkg/cmd/root"
)

type exitCode int

const (
	exitOK    exitCode = 0
	exitError exitCode = 1
)

func Run() exitCode {
	rootCmd := root.NewCmdRoot()

	if err := rootCmd.Execute(); err != nil {
		return exitError
	}

	return exitOK
}
