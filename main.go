package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/gookit/color"
)

type config struct {
	noStdout bool
	noStderr bool

	successTheme *color.Theme
	failureTheme *color.Theme

	successPattern *regexp.Regexp
	failurePattern *regexp.Regexp

	cmd     string
	cmdargs []string
}

func parseArgs() *config {
	conf := new(config)
	var (
		listThemes                     bool
		successTheme, failureTheme     string
		successPattern, failurePattern string
	)

	flag.BoolVar(&conf.noStdout, "no-stdout", false, "pass stdout through unchanged")
	flag.BoolVar(&conf.noStderr, "no-stderr", false, "pass stderr through unchanged")
	flag.BoolVar(&listThemes, "list-themes", false, "display the list of available themes")
	flag.StringVar(&successTheme, "success-theme", "success", "theme to use on success (use -list-themes flag to see options)")
	flag.StringVar(&failureTheme, "failure-theme", "error", "theme to use on failure (use -list-themes flag to see options)")
	flag.StringVar(&successPattern, "success-pattern", "(?i)\\b(success|pass|ok)\\b", "regular expression to identify success text")
	flag.StringVar(&failurePattern, "failure-pattern", "(?i)\\b(fail|failure|error)\\b", "regular expression to identify failure text")
	flag.Parse()

	if listThemes {
		printThemeList()
		os.Exit(0)
	}

	conf.successTheme = color.GetTheme(successTheme)
	conf.failureTheme = color.GetTheme(failureTheme)

	var err error
	conf.successPattern, err = regexp.Compile(successPattern)
	if err != nil {
		log.Fatal(err)
	}
	conf.failurePattern, err = regexp.Compile(failurePattern)
	if err != nil {
		log.Fatal(err)
	}

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
	fmt.Fprintf(os.Stderr, "%s [<flags>] <cmd> [--] [<cmdarg>, ...]\n", os.Args[0])
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
	var stdoutColor, stderrColor *colorizingWriter

	if conf.noStdout {
		cmd.Stdout = os.Stdout
	} else {
		stdoutColor = newColorizingWriter(os.Stdout, conf.successPattern, conf.failurePattern, conf.successTheme, conf.failureTheme)
		cmd.Stdout = stdoutColor
	}

	if conf.noStderr {
		cmd.Stderr = os.Stderr
	} else {
		stderrColor = newColorizingWriter(os.Stderr, conf.successPattern, conf.failurePattern, conf.successTheme, conf.failureTheme)
		cmd.Stderr = stderrColor
	}

	var exitErr *exec.ExitError
	err := cmd.Run()
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	} else if err != nil {
		log.Fatal(err)
	}
	return 0
}
