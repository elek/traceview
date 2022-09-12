package main

import tea "github.com/charmbracelet/bubbletea"

type Init struct {
	tea.Model
	initCmd tea.Cmd
}

func NewWithInit(model tea.Model, cmd tea.Cmd) *Init {
	return &Init{
		Model:   model,
		initCmd: cmd,
	}
}

var _ tea.Model = &Init{}

func (i *Init) Init() tea.Cmd {
	return i.initCmd
}
