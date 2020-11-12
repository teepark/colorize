package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/gookit/color"
)

type config struct {
	listThemes bool
	specs      []textColorSpec
}

const (
	failurePattern = "(?i)\\b(fail|failure|error)\\b"
	successPattern = "(?i)\\b(success|pass|ok)\\b"
)

func parseArgs() *config {
	conf := new(config)

	flag.BoolVar(&conf.listThemes, "list-themes", false, "display the list of available themes")
	flag.Var(
		(*specVar)(&conf.specs),
		"pattern",
		`'theme:regex' coloring pattern to apply.
Can be set more than once.
Defaults are:
    -pattern 'success:`+successPattern+`'
    -pattern 'error:`+failurePattern+`'
See -list-themes output for themes.`)
	flag.Parse()

	// default specs
	if len(conf.specs) == 0 {
		conf.specs = append(
			conf.specs,
			textColorSpec{
				pattern: regexp.MustCompile(failurePattern),
				theme:   color.GetTheme("error"),
			},
			textColorSpec{
				pattern: regexp.MustCompile(successPattern),
				theme:   color.GetTheme("success"),
			},
		)
	}

	return conf
}

func printThemeList() {
	for name, theme := range color.Themes {
		theme.Println(name)
	}
}

func usageAndExit() {
	fmt.Fprintf(os.Stderr, "%s [<flags>]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Usage = usageAndExit
	conf := parseArgs()
	if conf.listThemes {
		printThemeList()
		os.Exit(0)
	}
	io.Copy(newColorizingWriter(os.Stdout, conf.specs), os.Stdin)
}
