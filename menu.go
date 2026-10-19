package todo

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type Choice struct {
	Label    string
	OnSelect func() tea.Cmd
}

type Menu struct {
	Header  func() string
	Choices func() []Choice
	IsRoot  bool // the root menu has no "back"

	cursor int
}

// Text and Static adapt fixed values to Menu's function fields.
func Text(s string) func() string {
	return func() string {
		return s
	}
}
func Static(cs ...Choice) func() []Choice {
	return func() []Choice {
		return cs
	}
}

func (m *Menu) Init() tea.Cmd { return nil }

func (m *Menu) Update(msg tea.Msg) (Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	choices := m.Choices()
	m.cursor = clampCursor(m.cursor, len(choices)) // choices may have shrunk

	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		} else {
			m.cursor = len(choices) - 1
		}
	case "down", "j":
		if m.cursor < len(choices)-1 {
			m.cursor++
		} else {
			m.cursor = 0
		}
	case "b", "esc":
		if !m.IsRoot {
			return m, Pop()
		}
	case "enter":
		if len(choices) > 0 && choices[m.cursor].OnSelect != nil {
			return m, choices[m.cursor].OnSelect()
		}
	}
	return m, nil
}

func (m *Menu) View() string {
	var b strings.Builder
	b.WriteString(m.Header() + "\n\n")

	choices := m.Choices()
	cur := clampCursor(m.cursor, len(choices))
	for i, c := range choices {
		cursor := " "
		if i == cur {
			cursor = ">"
		}
		fmt.Fprintf(&b, "%s %s\n", cursor, c.Label)
	}

	if m.IsRoot {
		b.WriteString("\nPress ctrl+c to quit.\n")
	} else {
		b.WriteString("\nPress b to go back and ctrl+c to quit.\n")
	}
	return b.String()
}

func clampCursor(cursor, n int) int { return max(0, min(cursor, n-1)) }

// NewConfirm is just a Menu with two fixed choices.
func NewConfirm(question string, onYes func() tea.Cmd) *Menu {
	return &Menu{
		Header:  Text(question),
		Choices: Static(Choice{"yes", onYes}, Choice{"no", Pop}),
	}
}
