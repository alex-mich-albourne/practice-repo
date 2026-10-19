package todo

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func NewHomeScreen(s *Store) Screen {
	return &Menu{
		IsRoot: true,
		Header: Text("Welcome to Lists4u, what would you like to do?"),
		Choices: Static(
			Choice{"Create List", func() tea.Cmd { return Push(NewCreateListScreen(s)) }},
			Choice{"View Lists", func() tea.Cmd { return Push(NewViewListsScreen(s)) }},
			Choice{"Delete Lists", func() tea.Cmd { return Push(NewDeleteListsScreen(s)) }},
		),
	}
}

func NewCreateListScreen(s *Store) Screen {
	return NewPrompt("What do you want to name this list?", "Enter list name", func(name string) tea.Cmd {
		s.AddList(name)
		return Replace(NewViewListsScreen(s))
	})
}

// newListPicker is shared by "view" and "delete": same menu, different action.
func newListPicker(s *Store, header string, onPick func(i int) tea.Cmd) Screen {
	return &Menu{
		Header: Text(header),
		Choices: func() []Choice {
			choices := make([]Choice, len(s.Lists))
			for i, l := range s.Lists {
				choices[i] = Choice{l.Name, func() tea.Cmd { return onPick(i) }}
			}
			return choices
		},
	}
}

func NewViewListsScreen(s *Store) Screen {
	return newListPicker(s, "Choose a list to view:", func(i int) tea.Cmd {
		return Push(NewListEntriesScreen(s.Lists[i]))
	})
}

func NewDeleteListsScreen(s *Store) Screen {
	return newListPicker(s, "Choose a list to remove:", func(i int) tea.Cmd {
		q := fmt.Sprintf("Are you sure you want to delete list '%s'?", s.Lists[i].Name)
		return Push(NewConfirm(q, func() tea.Cmd {
			s.RemoveList(i)
			return PopToRoot()
		}))
	})
}

func NewListEntriesScreen(l *List) Screen {
	return &Menu{
		Header: func() string {
			var b strings.Builder
			fmt.Fprintf(&b, "Showing list entries for list '%s'", l.Name)
			for i, e := range l.Active() {
				fmt.Fprintf(&b, "\n\n%d - %s", i, e)
			}
			return b.String()
		},
		Choices: Static(
			Choice{"Add Entry", func() tea.Cmd { return Push(NewEntryTypeScreen(l)) }},
			Choice{"Remove Entry", func() tea.Cmd { return Push(NewRemoveEntryScreen(l)) }},
		),
	}
}

func NewEntryTypeScreen(l *List) Screen {
	choices := make([]Choice, len(EntryKinds))
	for i, kind := range EntryKinds {
		choices[i] = Choice{string(kind.Type), func() tea.Cmd {
			return Replace(NewWizard(kind.Fields, func(d Draft) tea.Cmd {
				l.AddEntry(kind.Build(d))
				return Pop()
			}))
		}}
	}
	return &Menu{Header: Text("Choose type of Entry:"), Choices: Static(choices...)}
}

func NewRemoveEntryScreen(l *List) Screen {
	return &Menu{
		Header: Text("Select entry to remove:"),
		Choices: func() []Choice {
			entries := l.Active()
			choices := make([]Choice, len(entries))
			for i, e := range entries {
				choices[i] = Choice{fmt.Sprintf("%d - %s", i, e), func() tea.Cmd {
					return Push(NewConfirm("Are you sure you want to delete this entry?", func() tea.Cmd {
						l.DeactivateListable(i)
						return PopN(2) // close the confirm and the picker
					}))
				}}
			}
			return choices
		},
	}
}
