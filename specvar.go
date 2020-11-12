package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gookit/color"
)

type specVar []textColorSpec

func (sv *specVar) String() string {
	return fmt.Sprintf("%+v\n", *sv)
}

func (sv *specVar) Set(value string) error {
	pair := strings.SplitN(value, ":", 2)
	if len(pair) < 2 {
		return fmt.Errorf("spec parameter malformed: '%s'", value)
	}

	themeName, reString := pair[0], pair[1]

	theme := color.GetTheme(themeName)
	if theme == nil {
		return fmt.Errorf("unrecognized theme name '%s'", themeName)
	}
	regex, err := regexp.Compile(reString)
	if err != nil {
		return err
	}

	*sv = append(*sv, textColorSpec{pattern: regex, theme: theme})
	return nil
}
