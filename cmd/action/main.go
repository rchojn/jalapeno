package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/futurice/jalapeno/internal/cli"
)

var (
	// https://goreleaser.com/cookbooks/using-main.version/
	version string
)

func main() {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		checkErr(errors.New("this image only works on Github Actions"))
	}

	filename := os.Getenv("GITHUB_OUTPUT")
	if filename == "" {
		checkErr(errors.New("GITHUB_OUTPUT environment variable not set"))
	}

	output, err := os.OpenFile(filename, os.O_APPEND, 0644)
	checkErr(err)
	defer output.Close()

	cmd := cli.NewRootCmd(version)
	err = cmd.ExecuteContext(context.Background())

	exitCode, isExitCodeSet := cmd.Context().Value(cli.ExitCodeContextKey{}).(int)
	if !isExitCodeSet {
		if err == nil {
			exitCode = 0
		} else {
			exitCode = 1
		}
	}
	fmt.Fprintf(output, "exit-code=%d\n", exitCode)
	os.Exit(exitCode)
}

func checkErr(err error) {
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
