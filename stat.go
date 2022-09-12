package main

import (
	"fmt"
	ui "github.com/elek/bubbles"
)

type Stat struct {
	*ui.Text
	trace *LoadedTrace
	count int
}

func (s Stat) Render() string {
	out := ""
	out += fmt.Sprintf("Number of spans: %d\n", len(s.trace.AllTraces))
	out += fmt.Sprintf("Orphans:         %d", len(s.trace.Orphans.Children))
	return out
}

func NewStat(trace *LoadedTrace) Stat {
	s := Stat{
		trace: trace,
	}
	s.Text = ui.NewText(s.Render)
	return s
}
