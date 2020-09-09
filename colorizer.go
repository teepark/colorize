package main

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/gookit/color"
)

type textColorSpec struct {
	pattern *regexp.Regexp
	theme   *color.Theme
}

func (tcs textColorSpec) colorize(raw string) string {
	var b strings.Builder
	matches := tcs.pattern.FindAllStringIndex(raw, -1)
	latest := 0
	for _, match := range matches {
		b.WriteString(raw[latest:match[0]])
		b.WriteString(tcs.theme.Sprint(raw[match[0]:match[1]]))
		latest = match[1]
	}
	b.WriteString(raw[latest:len(raw)])
	return b.String()
}

type colorizingWriter struct {
	w     io.Writer
	specs []textColorSpec
}

func newColorizingWriter(w io.Writer, specs []textColorSpec) *colorizingWriter {
	return &colorizingWriter{w: w, specs: specs}
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

		for _, spec := range cw.specs {
			line = spec.colorize(line)
		}

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
