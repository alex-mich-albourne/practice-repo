package todo

import tea "charm.land/bubbletea/v2"

type App struct {
	stack []Screen
}

func NewApp() *App {
	return &App{stack: []Screen{NewHomeScreen(&Store{})}}
}

func (a *App) top() Screen {
	return a.stack[len(a.stack)-1]
}

func (a *App) Init() tea.Cmd {
	return a.top().Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case pushMsg:
		a.stack = append(a.stack, msg.screen)
		return a, msg.screen.Init()
	case replaceMsg:
		a.stack[len(a.stack)-1] = msg.screen
		return a, msg.screen.Init()
	case popMsg:
		n := min(msg.n, len(a.stack)-1) // never pop the root
		a.stack = a.stack[:len(a.stack)-n]
		return a, nil
	case popToRootMsg:
		a.stack = a.stack[:1]
		return a, nil
	}

	next, cmd := a.top().Update(msg)
	a.stack[len(a.stack)-1] = next
	return a, cmd
}

func (a *App) View() tea.View {
	return tea.NewView(a.top().View())
}
