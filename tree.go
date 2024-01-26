package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ui "github.com/elek/bubbles"
	"github.com/montanaflynn/stats"
	"time"
)

type NavigateTo struct {
	SpanID string
}

func NewTreePane(root *TreeSpan) tea.Model {

	master := &TreeMaster{
		TreeList: ui.NewTreeList[*TreeSpan, *TreeSpan](
			root,
			nil,
			func(l *TreeSpan) []*TreeSpan {
				return l.Children
			},
			func(l *TreeSpan) (*TreeSpan, bool) {
				return l, len(l.Children) > 0
			}),
	}
	master.Render = func(span *TreeSpan) string {
		return RenderSpan(master.CurrentLevel(), span)
	}

	detail := ui.NewDetail[*TreeSpan](func(span *TreeSpan) string {
		return RenderSpanInfo(master.CurrentLevel(), span)
	})
	detail.ChangeStyle(func(orig lipgloss.Style) lipgloss.Style {
		return orig.Border(lipgloss.NormalBorder()).BorderForeground(ui.White)
	})
	mw := ui.WithBorder(ui.NewHeader(RenderHeader(), master))

	dual := ui.Horizontal{}
	dual.Add(mw, ui.RemainingSize())
	dual.Add(detail, ui.FixedSize(50))

	return &dual
}

func RenderHeader() string {
	return fmt.Sprintf("%-17s %19s %11s %17s %17s %s", "ID", "TIME", "SERVICE", "DURATION", "TIMEBOX", "NAME")
}
func RenderSpan(current *TreeSpan, span *TreeSpan) string {

	startOffset := span.StartTime.Sub(current.StartTime).Microseconds()

	canceled := func(style lipgloss.Style) lipgloss.Style {
		if span.HasTag("status", "canceled") {
			return style.Foreground(Red)
		}
		return style
	}

	size := func(size int, style lipgloss.Style) lipgloss.Style {
		return style.MaxWidth(size).Width(size).PaddingRight(1)
	}
	l := NewLineBuilder(300)
	def := lipgloss.NewStyle()
	l.AddSegment(span.SpanID, size(18, def))
	l.AddSegment(timeFormat(int(startOffset)), size(20, Colored(Blue)).Align(lipgloss.Right))
	l.AddSegment(span.Process.ServiceName, size(12, def))
	l.AddSegment(timeFormat(span.Duration), size(18, Colored(Green)).Align(lipgloss.Right))
	l.AddSegment(timeFormat(span.Timebox), size(18, Colored(Yellow)).Align(lipgloss.Right))
	l.AddSegment(span.OperationName, size(200, canceled(def)))

	return l.String()
}

func RenderSpanInfo(current *TreeSpan, selected *TreeSpan) string {
	out := ""
	out += "Parent:\n"
	if current.Process != nil {
		out += lipgloss.NewStyle().Bold(true).Render(current.Process.ServiceName) + "\n"
	}
	out += Colored(Red).Bold(false).Render(current.OperationName) + "\n"
	out += Colored(ui.White).Bold(false).Render(current.SpanID) + "\n"
	out += fmt.Sprintf("%-16s: %16s\n", "Duration:", timeFormat(current.Duration))
	out += fmt.Sprintf("%-16s: %16s\n", "Start:", current.StartTime.Format(time.RFC3339))

	fcount := 0
	fsum := 0
	fsumR := 0
	for _, c := range current.Children {
		fcount++
		fsum += c.Duration
		fsumR += c.Timebox
	}
	out += "\n"
	out += "Children...\n"
	out += fmt.Sprintf("%-16s: %16d\n", "  Count", fcount)
	out += fmt.Sprintf("%-16s: %s\n", "  Sum", Colored(Green).Width(16).Align(lipgloss.Right).Render(timeFormat(fsum)))
	out += fmt.Sprintf("%-16s: %s\n", "  Timebox Sum", Colored(Yellow).Width(16).Align(lipgloss.Right).Render(timeFormat(fsumR)))

	if selected != nil {
		out += "\nThis:\n"
		out += Colored(Red).Bold(true).Render(selected.OperationName) + "\n"
		out += fmt.Sprintf("%-16s: %s\n", "Duration", Colored(Green).Width(16).Align(lipgloss.Right).Render(timeFormat(selected.Duration)))
		out += fmt.Sprintf("%-16s: %s\n", "Timebox", Colored(Yellow).Width(16).Align(lipgloss.Right).Render(timeFormat(selected.Timebox)))
		out += "\n"
		count := 0
		sum := 0
		sumR := 0
		data := []float64{}
		for _, c := range current.Children {
			if c.OperationName == selected.OperationName {
				count++
				sum += c.Duration
				sumR += c.Timebox
				data = append(data, float64(c.Duration))
			}
		}

		if count > 1 {
			out += "\nAll:\n"
			out += Colored(Red).Bold(true).Render(selected.OperationName) + "\n"
			out += fmt.Sprintf("%-16s: %16d\n", "Count", count)
			out += fmt.Sprintf("%-16s: %16s\n", "Sum", timeFormat(sum))
			out += fmt.Sprintf("%-16s: %16s\n", "Avg", timeFormat(sum/count))
			out += fmt.Sprintf("%-16s: %16s\n", "Timebox Sum", timeFormat(sumR))
			out += fmt.Sprintf("%-16s: %16s\n", "Timebox Avg", timeFormat(sumR/count))
			if count > 5 {
				out += fmt.Sprintf("%-16s: %16s\n", "r99", perc(data, 99))
				out += fmt.Sprintf("%-16s: %16s\n", "r95", perc(data, 95))
				out += fmt.Sprintf("%-16s: %16s\n", "r90", perc(data, 90))
				out += fmt.Sprintf("%-16s: %16s\n", "r75", perc(data, 75))
				out += fmt.Sprintf("%-16s: %16s\n", "r50", perc(data, 50))
			}
		}
		out += "\n"
		for _, t := range selected.Tags {
			out += fmt.Sprintf("%-16s: %v\n", t.Key, t.Value)
		}

		if selected.Process != nil {
			out += "\nProcess:\n"
			out += fmt.Sprintf("%s (# %s)\n", selected.Process.ServiceName, selected.Process.ID)
			for _, tags := range selected.Process.Tags {
				out += fmt.Sprintf("  %s=%s\n", tags.Key, tags.Value)
			}
		}
	}
	return out
}

func perc(data []float64, f float64) string {
	r, err := stats.Percentile(data, f)
	if err != nil {
		return err.Error()
	}
	return timeFormat(int(r))
}

type TreeMaster struct {
	*ui.TreeList[*TreeSpan, *TreeSpan]
}

func (t *TreeMaster) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, c := t.TreeList.Update(msg)
	t.TreeList = m.(*ui.TreeList[*TreeSpan, *TreeSpan])
	switch msg := msg.(type) {
	case NavigateTo:
		t.TreeList.NavigateTo(func(span *TreeSpan) bool {
			return span.SpanID == msg.SpanID
		})
	}
	t.TreeList = m.(*ui.TreeList[*TreeSpan, *TreeSpan])
	return t, c
}
