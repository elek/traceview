package main

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type Stack struct {
	Path   string `arg:""`
	Filter string `arg:""`
}

func (s Stack) Run() error {
	trace, err := load(s.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, span := range trace.Tree.RootSpans.Children {
		s.printRecursive([]*TreeSpan{span})
	}
	return nil
}

func (s Stack) printRecursive(span []*TreeSpan) {
	curr := span[len(span)-1]
	if strings.Contains(curr.OperationName, s.Filter) {
		for _, sp := range span {

			fmt.Printf("%10s %10s   %s\n", timeFormat(sp.Duration), timeFormat(sp.Timebox), sp.OperationName)
		}
		fmt.Println()
		return
	}
	for _, c := range curr.Children {
		s.printRecursive(append(append([]*TreeSpan{}, span...), c))
	}
}
