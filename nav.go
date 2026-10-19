package todo

import tea "charm.land/bubbletea/v2"

type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
}

type pushMsg struct {
	screen Screen
}
type replaceMsg struct {
	screen Screen
}
type popMsg struct {
	n int
}
type popToRootMsg struct{}

func Push(s Screen) tea.Cmd {
	return func() tea.Msg {
		return pushMsg{s}
	}
}
func Replace(s Screen) tea.Cmd {
	return func() tea.Msg {
		return replaceMsg{s}
	}
}
func Pop() tea.Cmd {
	return PopN(1)
}
func PopN(n int) tea.Cmd {
	return func() tea.Msg {
		return popMsg{n}
	}
}
func PopToRoot() tea.Cmd {
	return func() tea.Msg {
		return popToRootMsg{}
	}
}
