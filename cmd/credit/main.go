/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/ffalor/credit/pkg/cli"
)

func main() {
	os.Exit(int(cli.Run()))
}
