package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ui "github.com/elek/bubbles"
	"github.com/zeebo/errs/v2"
	"log"
	"os"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalf("%++v", err)
	}
}

func run() error {
	l, _ := tea.LogToFile("/tmp/tea.log", "traceview")
	defer l.Close()
	trace, err := load(os.Args[1])
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
