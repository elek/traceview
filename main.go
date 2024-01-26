package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ui "github.com/elek/bubbles"
	"github.com/zeebo/errs/v2"
	"os"
)

func main() {
	traceview := Traceview{}
	kCtx := kong.Parse(&traceview)
	err := kCtx.Run()
	kCtx.FatalIfErrorf(err)
}

type Traceview struct {
	Grep    Grep      `cmd:"" `
	Process Processor `cmd:"" `
	View    UI        `cmd:"" default:"withargs"`
	Stack   Stack     `cmd:""`
}

type UI struct {
	TraceFile string `arg:""`
}

func (u UI) Run() error {
	l, _ := tea.LogToFile("/tmp/tea.log", "traceview")
	defer l.Close()
	trace, err := load(u.TraceFile)
	if err != nil {
		return errs.Wrap(err)
	}

	var init tea.Cmd
	if len(os.Args) > 2 {
		init = ui.AsCommand(NavigateTo{
			SpanID: os.Args[2],
		})
	}

	stat := NewStat(trace)
	stat.Style = stat.Style.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#EFEFEF"))
	program := tea.NewProgram(NewWithInit(ui.NewKillable(
		ui.NewTabs(
			ui.Tab{
				Name:  "tree",
				Model: NewTreePane(trace.Tree.RootSpans),
				Key:   "1",
			},
			ui.Tab{
				Name:  "orphans",
				Model: NewTreePane(trace.Orphans),
				Key:   "2",
			},
			ui.Tab{
				Name:  "operation",
				Model: NewOperationPane(trace.Tree.RootSpans),
				Key:   "3",
			},
			ui.Tab{
				Name:  "stat",
				Model: stat,
				Key:   "4",
			},
		)), init))

	if err := program.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}
