package main

import (
	"github.com/charmbracelet/lipgloss"
)

type LineBuilder struct {
	width int
	used  int
	out   string
}

func NewLineBuilder(width int) *LineBuilder {
	return &LineBuilder{
		width: width,
	}
}

func (l *LineBuilder) AddSegment(content string, style lipgloss.Style) {
	width := style.GetMaxWidth()
	if l.used >= l.width {
		return
	}
	if width == 0 {
		width = 1000
	}
	if l.width-l.used < width {
		width = l.width - l.used
		style = style.MaxWidth(width).MaxWidth(width)
	}
	realWidth := width - style.GetHorizontalPadding()
	if realWidth < len(content) {
		content = content[0:realWidth]
	}

	rendered := style.Render(content)
	l.out += rendered
	renderedSize := lipgloss.Width(rendered)
	l.used += renderedSize

}

func (l *LineBuilder) String() string {
	return l.out
}
