package todo

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Draft collects answers while an entry is being built.
type Draft struct {
	Title       string
	Description string
	When        time.Time
	Priority    Priority
	Online      bool
	Location    string
	Platform    string
}

// A Field is one question. Its screen writes the answer into the Draft and
// then returns nextField so the Wizard moves on.
type Field struct {
	Skip   func(d *Draft) bool // optional: skip this question based on earlier answers
	Screen func(d *Draft) Screen
}

type nextFieldMsg struct{}

func nextField() tea.Msg { return nextFieldMsg{} }

// Wizard runs a sequence of Fields, showing one child screen at a time.
type Wizard struct {
	draft  *Draft
	fields []Field
	index  int
	child  Screen
	onDone func(Draft) tea.Cmd
}

func NewWizard(fields []Field, onDone func(Draft) tea.Cmd) *Wizard {
	return &Wizard{draft: &Draft{}, fields: fields, index: -1, onDone: onDone}
}

func (w *Wizard) Init() tea.Cmd { return w.advance() }

func (w *Wizard) advance() tea.Cmd {
	for w.index++; w.index < len(w.fields); w.index++ {
		f := w.fields[w.index]
		if f.Skip == nil || !f.Skip(w.draft) {
			w.child = f.Screen(w.draft)
			return w.child.Init()
		}
	}
	return w.onDone(*w.draft)
}

func (w *Wizard) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if _, ok := msg.(nextFieldMsg); ok {
		return w, w.advance()
	}
	var cmd tea.Cmd
	w.child, cmd = w.child.Update(msg)
	return w, cmd
}

func (w *Wizard) View() string {
	if w.child == nil {
		return ""
	}
	return w.child.View()
}
