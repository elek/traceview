package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ui "github.com/elek/bubbles"
	"sort"
)

type Operation struct {
	*ui.FilterableList[OperationStat]
}

var _ tea.Model = &Operation{}

func RenderStat(stat OperationStat) string {
	return fmt.Sprintf("%6d %16s %16s %30s", stat.Count, timeFormat(stat.Max), timeFormat(stat.Min), stat.Name)
}

func NewOperationPane(root *TreeSpan) tea.Model {
	l := ui.NewList([]*TreeSpan{}, RenderStatTrace)
	l.ChangeStyle(func(orig lipgloss.Style) lipgloss.Style {
		return orig.Border(lipgloss.NormalBorder()).BorderForeground(ui.White)
	})

	detail := OperationInstance{
		List: l,
	}

	info := ui.NewDetail[OperationStat](func(stat OperationStat) string {
		out := ""

		var values []float64
		sum := 0
		errors := 0
		for _, s := range stat.Spans {
			if !s.HasTag("status", "canceled") {
				values = append(values, float64(s.Duration))
				sum += s.Duration
			} else {
				errors++
			}
		}

		out += fmt.Sprintf("%-10s: %d\n", "Count", len(stat.Spans))
		out += fmt.Sprintf("%-10s: %d\n", "Error", errors)
		out += fmt.Sprintf("%-10s: %s\n", "r99", perc(values, 99))
		out += fmt.Sprintf("%-10s: %s\n", "r90", perc(values, 90))
		out += fmt.Sprintf("%-10s: %s", "r75", perc(values, 75))
		return out
	})
	info.ChangeStyle(func(orig lipgloss.Style) lipgloss.Style {
		return orig.Border(lipgloss.NormalBorder()).BorderForeground(ui.White)
	})

	v := ui.Vertical{}
	v.Add(info, ui.FixedSize(6))
	v.Add(&detail, ui.RemainingSize())

	h := ui.Horizontal{}

	master := NewOperation(root)
	master.ChangeStyle(func(orig lipgloss.Style) lipgloss.Style {
		return orig.Border(lipgloss.NormalBorder()).BorderForeground(ui.White)
	})

	h.Add(master, ui.RatioSize(2))
	h.Add(&v, ui.RemainingSize())

	fg := ui.NewFocusGroup(&h)
	fg.Add(master)
	fg.Add(&detail)
	return fg

}

func collectFromTree(base *TreeSpan, l map[string]*OperationStat) {
	for _, span := range base.Children {
		if _, found := l[span.OperationName]; !found {
			l[span.OperationName] = &OperationStat{
				Name:  span.OperationName,
				Min:   -1,
				Spans: make([]*TreeSpan, 0),
			}
		}
		l[span.OperationName].Spans = append(l[span.OperationName].Spans, span)
		l[span.OperationName].Count++
		if l[span.OperationName].Max < span.Duration {
			l[span.OperationName].Max = span.Duration
		}
		if l[span.OperationName].Min == -1 || l[span.OperationName].Min > span.Duration {
			l[span.OperationName].Min = span.Duration
		}
		collectFromTree(span, l)
	}

}
func NewOperation(root *TreeSpan) *Operation {
	o := Operation{}
	l := make(map[string]*OperationStat)
	collectFromTree(root, l)

	stat := make([]OperationStat, 0)
	for _, s := range l {
		stat = append(stat, *s)
	}

	sort.Slice(stat, func(i, j int) bool {
		return stat[i].Count > stat[j].Count
	})

	o.FilterableList = ui.NewFilterableList[OperationStat](stat, RenderStat, nil)
	return &o
}

type OperationStat struct {
	Name  string
	Count int
	Min   int
	Max   int
	Spans []*TreeSpan
}

type OperationInstance struct {
	*ui.List[*TreeSpan]
}

func (o *OperationInstance) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ui.FocusedItemMsg[OperationStat]:
		o.List.SetContent(msg.Item.Spans)
		o.List.Reset()
	case tea.KeyMsg:
		if o.Focused {
			switch msg.String() {
			case "enter":
				selected := o.List.Selected()
				return o, ui.AsCommand(ui.ActivateTabMsg{Name: "tree", Msg: NavigateTo{
					SpanID: selected.SpanID,
				}})
			}
		}
	}
	m, c := o.List.Update(msg)
	o.List = m.(*ui.List[*TreeSpan])
	return o, c
}

func RenderStatTrace(instance *TreeSpan) string {
	return fmt.Sprintf("%10s %20s %16s %16s", instance.SpanID, timeFormat(instance.Duration), timeFormat(instance.Timebox), instance.Process.ServiceName)
}
