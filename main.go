package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/gookit/color"
)

type config struct {
	noStdout bool
	noStderr bool

	successTheme *color.Theme
	failureTheme *color.Theme

	cmd     string
	cmdargs []string
}

func parseArgs() *config {
	conf := new(config)
	var (
		listThemes                 bool
		successTheme, failureTheme string
	)

	flag.BoolVar(&conf.noStdout, "no-stdout", false, "pass stdout through unchanged")
	flag.BoolVar(&conf.noStderr, "no-stderr", false, "pass stderr through unchanged")
	flag.BoolVar(&listThemes, "list-themes", false, "display the list of available themes")
	flag.StringVar(&successTheme, "success-theme", "success", "theme to use on success (use -list-themes flag to see options)")
	flag.StringVar(&failureTheme, "failure-theme", "danger", "theme to use on failure (use -list-themes flag to see options)")
	flag.Parse()

	if listThemes {
		printThemeList()
		os.Exit(0)
	}

	conf.successTheme = color.GetTheme(successTheme)
	conf.failureTheme = color.GetTheme(failureTheme)

	args := flag.Args()
	if len(args) == 0 || conf.successTheme == nil || conf.failureTheme == nil {
		usageAndExit()
	}

	conf.cmd = args[0]
	conf.cmdargs = args[1:]

	return conf
}

func printThemeList() {
	for name, theme := range color.Themes {
		theme.Println(name)
	}
}

func usageAndExit() {
	fmt.Fprintf(os.Stderr, "%s [<flags>] <cmd> [<cmdarg>, ...]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Usage = usageAndExit
	conf := parseArgs()
	os.Exit(runCommand(conf))
}

func runCommand(conf *config) int {
	cmd := exec.CommandContext(context.Background(), conf.cmd, conf.cmdargs...)
	var stdoutBuf, stderrBuf *bytes.Buffer

	if conf.noStdout {
		cmd.Stdout = os.Stdout
	} else {
		stdoutBuf = new(bytes.Buffer)
		cmd.Stdout = stdoutBuf
	}
	if conf.noStderr {
		cmd.Stderr = os.Stderr
	} else {
		stderrBuf = new(bytes.Buffer)
		cmd.Stderr = stderrBuf
	}

	err := cmd.Run()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		log.Fatal(err)
	}

	exitCode := cmd.ProcessState.ExitCode()

	var theme *color.Theme
	if exitCode == 0 {
		theme = conf.successTheme
	} else {
		theme = conf.failureTheme
	}

	if !conf.noStdout {
		io.WriteString(os.Stdout, theme.Sprint(stdoutBuf.String()))
	}
	if !conf.noStderr {
		io.WriteString(os.Stderr, theme.Sprint(stderrBuf.String()))
	}

	return exitCode
}
