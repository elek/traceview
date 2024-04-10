package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

type Processor struct {
	Path        string `arg:""`
	FilterFile  string `arg:""`
	ProcessTags bool   `help:"print out process tags"`
	FullPath    bool   `help:"print out full path of the input file"`
}

func (a Processor) Run() error {
	stat, err := os.Stat(a.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	if stat.IsDir() {
		entries, err := os.ReadDir(a.Path)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), "json") {
				err := a.processFile(filepath.Join(a.Path, e.Name()))
				if err != nil {
					return err
				}
				fmt.Println()
			}
		}
	} else {
		err := a.processFile(a.Path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a Processor) processFile(path string) error {
	trace, err := load(path)
	if err != nil {
		return errors.WithStack(err)
	}
	tags := ""
	var filters []SpanFilter
	lines, err := os.ReadFile(a.FilterFile)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, line := range strings.Split(string(lines), "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "":
			continue
		case "@root":
			filters = append(filters, FilterRoot())
		default:
			filters = append(filters, FilterByName(line))
		}
	}

	return trace.Tree.RootSpans.Walk(func(parent *TreeSpan, span *TreeSpan) (bool, error) {
		for _, f := range filters {
			if f(parent, span) {
				if a.ProcessTags && tags == "" {
					k := []string{}
					for _, t := range span.Process.Tags {
						k = append(k, fmt.Sprintf("%s=%s", t.Key, t.Value))
					}
					tags = " " + strings.Join(k, ",")
				}
				var f string
				if a.FullPath {
					f = path
				} else {
					f = filepath.Base(path)
				}

				fmt.Printf("%s %s%s %d\n", f, span.OperationName, tags, span.Duration)
			}
		}
		return true, nil
	})
}

type SpanFilter func(parent *TreeSpan, span *TreeSpan) bool

func FilterByName(pattern string) SpanFilter {
	return func(parent *TreeSpan, span *TreeSpan) bool {
		if pattern[len(pattern)-1] == '$' {
			return strings.HasSuffix(span.OperationName, pattern[:len(pattern)-1])
		} else {
			return strings.Contains(span.OperationName, pattern)
		}
	}
}

func FilterRoot() SpanFilter {
	return func(parent *TreeSpan, span *TreeSpan) bool {
		return parent.OperationName == ""
	}
}
