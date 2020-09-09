package main

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/gookit/color"
)

func colorize(raw string, theme *color.Theme, re *regexp.Regexp) string {
	var b strings.Builder
	matches := re.FindAllStringIndex(raw, -1)
	latest := 0
	for _, match := range matches {
		b.WriteString(raw[latest:match[0]])
		b.WriteString(theme.Sprint(raw[match[0]:match[1]]))
		latest = match[1]
	}
	b.WriteString(raw[latest:len(raw)])
	return b.String()
}

type colorizingWriter struct {
	w   io.Writer

	successPattern *regexp.Regexp
	failurePattern *regexp.Regexp

	successTheme *color.Theme
	failureTheme *color.Theme
}

func newColorizingWriter(w io.Writer, successPattern, failurePattern *regexp.Regexp, successTheme, failureTheme *color.Theme) *colorizingWriter {
	return &colorizingWriter{
		w:   w,

		successPattern: successPattern,
		failurePattern: failurePattern,

		successTheme: successTheme,
		failureTheme: failureTheme,
	}
}

func (cw *colorizingWriter) Write(data []byte) (nn int, err error) {
	// colorizingWriter must implement the io.Writer interface to type check,
	// but if we implement io.ReaderFrom then io.Copy() will only use ReadFrom().
	// It's much harder to implement reentrant writes so I'm not bothering.
	panic("no implementation")
}

func (cw *colorizingWriter) ReadFrom(r io.Reader) (int64, error) {
	buf := bufio.NewReader(r)
	var written int64
	for {
		line, rderr := buf.ReadString('\n')
		written += int64(len(line))

		line = colorize(line, cw.failureTheme, cw.failurePattern)
		line = colorize(line, cw.successTheme, cw.successPattern)

		_, wrerr := io.WriteString(cw.w, line)
		if wrerr != nil {
			return written, wrerr
		}

		if rderr != nil {
			if errors.Is(rderr, io.EOF) {
				rderr = nil
			}
			return written, rderr
		}
	}
}
