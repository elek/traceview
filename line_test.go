package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"testing"
)

func TestRenderLine(t *testing.T) {
	l := NewLineBuilder(20)
	st := lipgloss.NewStyle().Bold(true)
	l.AddSegment("123456789 123456789 xxx", st.Width(10).MaxWidth(10).PaddingRight(1).Background(Blue))
	l.AddSegment("123456789 123456789 xxx", st.Background(Red))
	fmt.Println(l.String())
	fmt.Println(l.String())
}
