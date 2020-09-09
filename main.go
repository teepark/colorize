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

	specs []textColorSpec

	cmd     string
	cmdargs []string
}

const (
	successPattern = "(?i)\\b(fail|failure|error|no)\\b"
	failurePattern = "(?i)\\b(success|pass|ok|yes)\\b"
)

func parseArgs() *config {
	conf := new(config)
	var listThemes bool

	flag.BoolVar(&conf.noStdout, "no-stdout", false, "pass stdout through unchanged")
	flag.BoolVar(&conf.noStderr, "no-stderr", false, "pass stderr through unchanged")
	flag.BoolVar(&listThemes, "list-themes", false, "display the list of available themes")
	flag.Var(
		(*specVar)(&conf.specs),
		"pattern",
		`'theme:regex' coloring pattern to apply.
Can be set more than once.
Defaults are:
    -pattern 'success:`+failurePattern+`'
    -pattern 'error:`+successPattern+`'
See -list-themes output for themes.
`)
	flag.Parse()

	if listThemes {
		printThemeList()
		os.Exit(0)
	}

	// default specs
	if len(conf.specs) == 0 {
		conf.specs = append(
			conf.specs,
			textColorSpec{
				pattern: regexp.MustCompile(successPattern),
				theme:   color.GetTheme("error"),
			},
			textColorSpec{
				pattern: regexp.MustCompile(failurePattern),
				theme:   color.GetTheme("success"),
			},
		)
	}

	args := flag.Args()
	if len(args) == 0 {
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
		stdoutColor = newColorizingWriter(os.Stdout, conf.specs)
		cmd.Stdout = stdoutColor
	}

	if conf.noStderr {
		cmd.Stderr = os.Stderr
	} else {
		stderrColor = newColorizingWriter(os.Stderr, conf.specs)
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
