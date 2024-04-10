package main

import (
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

type Grep struct {
	Path        string `arg:""`
	Pattern     string `arg:""`
	ProcessTags bool   `help:"print out process tags"`
	FullPath    bool   `help:"print out full path of the input file"`
}

func (a Grep) Run() error {
	return ProcessPath(a.Path, func(f string) error {
		return a.processFile(f)
	})
}

func (a Grep) processFile(path string) error {
	trace, err := load(path)
	if err != nil {
		return errors.WithStack(err)
	}
	tags := ""
	return trace.Tree.RootSpans.Walk(func(parent *TreeSpan, span *TreeSpan) (bool, error) {
		if match(span.OperationName, a.Pattern) {
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
		return true, nil
	})
}

func match(name string, pattern string) bool {
	if pattern[len(pattern)-1] == '$' {
		return strings.HasSuffix(name, pattern[:len(pattern)-1])
	} else {
		return strings.Contains(name, pattern)
	}
}
