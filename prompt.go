package todo

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// Prompt is a reusable "question + text input" screen.
type Prompt struct {
	Question string
	OnSubmit func(value string) tea.Cmd
	Validate func(value string) error // optional

	input textinput.Model
	err   error
}

func NewPrompt(question, placeholder string, onSubmit func(string) tea.Cmd) *Prompt {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 50
	ti.SetWidth(30)
	return &Prompt{Question: question, OnSubmit: onSubmit, input: ti}
}

func (p *Prompt) Init() tea.Cmd { return p.input.Focus() }

func (p *Prompt) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			return p, Pop()
		case "enter":
			v := p.input.Value()
			if p.Validate != nil {
				if p.err = p.Validate(v); p.err != nil {
					return p, nil
				}
			}
			return p, p.OnSubmit(v)
		}
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return p, cmd
}

func (p *Prompt) View() string {
	s := p.Question + "\n\n" + p.input.View()
	if p.err != nil {
		s += "\n\n" + p.err.Error()
	}
	return s + "\n\nPress esc to go back."
}
