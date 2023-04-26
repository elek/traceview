package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

type Grep struct {
	Pattern     string `arg:""`
	Path        string `arg:""`
	ProcessTags bool   `help:"print out process tags"`
	FullPath    bool   `help:"print out full path of the input file"`
}

func (a Grep) Run() error {
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
