# Making the Screens Modular

This doc has two parts:

1. **The general pattern.** How to structure a multi-screen terminal UI (or any UI) so each screen is self-contained.
2. **Applying it to Lists4u.** Concrete code for this repo, how each old screen maps to the new design, and a migration path.

All the code in Part 2 has been compiled against this repo's `bubbletea/v2 v2.1.0` and `bubbles/v2 v2.2.1`, passes `go vet`, and was driven end to end in a scratch copy.

---

## Where the code is today

Splitting the app into screens was the right instinct. A `Screen` enum is a real improvement over a pile of booleans. The problem is that **each screen's behavior is spread across five places**:

| Concern | Where it lives |
|---|---|
| Which screens exist | `configs.go` (`Screen` enum) |
| What a screen shows | `todo_logic.go` `UpdateChoices` (header/choices) **and** `todo_model.go` `View` (special cases for input screens and the entries screen) |
| What `enter` does | `todo_logic.go` `determineSelection` **and** `determineNextField` **and** `key_bindings.go` (the `enter` if-chain) |
| What `b` / `ctrl+c` do | `key_bindings.go` (lists of screens) |
| Screen-specific state | `Model` fields like `listPendingRemoval`, `entryPendingRemoval`, `currentlySelectedList`, `currentEntry`, `inputPrefix` |

So adding or changing one screen means editing about five `switch`/`if` chains and adding fields to a shared `Model` that every other screen can see. That's the "shotgun surgery" smell. Some symptoms already in the code:

- **"Back" is hand-coded per screen.** `key_bindings.go:30-44` lists screens by name to decide where `b` goes. Lines 40-42 can never run, because that branch excludes `ViewListEntriesScreen` and then checks for it.
- **The entry wizard is a hard-coded state machine.** `determineNextField` encodes "ToDo asks title → description → date → priority" as nested switches on `screen` × `entryType`. Adding a new entry type touches every case.
- **Shared mutable state leaks between flows:**
  - `currentEntry` is never reset. Its `priority` is incremented on each ToDo (`todo_logic.go:35`), so the second ToDo gets a different priority than the first.
  - The priority menu's cursor is never read, so the user's choice is ignored.
  - Meetings and events always get priority `0`.
- **Edge cases fall through the cracks:**
  - Pressing `enter` on *View Lists* with no lists panics (`m.lists[0]` in `UpdateChoices`).
  - `ctrl+c` on an entry prompt quits the app even though the footer says "close".
  - A bad date is silently parsed as the zero time (`key_bindings.go:60`).
- `go vet` flags three copies of `List` (which contains a `sync.Mutex`), because lists are stored as `[]List` by value.

None of these are hard bugs to fix one by one. They exist because the structure makes them easy to write.

---

# Part 1: The general pattern

The core idea is to stop modeling a screen as a *value* (an enum) and start modeling it as an *object*. Each screen owns its state, its key handling, and its rendering. This is the classic **State pattern**. In Elm-architecture frameworks like Bubble Tea, it's usually called **nested models** or **components**.

The approach has five pieces.

### 1. A `Screen` interface

```go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View() string
}
```

This mirrors `tea.Model`, but returns `Screen` and `string`. Each concrete screen is a struct that holds only *its own* state: a cursor, a text input, "which list am I showing". Nothing global.

### 2. One root model that routes messages

Only one type implements `tea.Model`. It holds the active screen(s) and forwards each message to the top one. It also handles truly global concerns like `ctrl+c` to quit and window resizing. The root never contains screen-specific logic.

### 3. Navigation as messages, with a stack

Screens don't set `m.screen = X`. They return a **command** that asks for navigation:

```go
return Push(NewSomeScreen(...))   // go forward
return Pop()                      // go back
return Replace(NewOther(...))     // swap the current screen
return PopToRoot()                // jump home
```

The root keeps a **stack** (`[]Screen`). This gives you "back" for free: it means "pop". You no longer maintain a table of "from screen X, back goes to Y". It also means screens don't know who opened them, so the same screen can be reused from several places.

### 4. Reusable primitives instead of one-off screens

Most TUI screens are one of a few shapes:

- **Menu**: a header and a list of choices; each choice has an action.
- **Prompt**: a question and a text input; submitting runs an action. Validation is optional.
- **Confirm**: a yes/no question. This is just a Menu with two choices.

Build these once and make them configurable with *callbacks*. A "screen" then becomes a small constructor function that wires a primitive to some data and actions. Fifteen hand-written screens become three components and a handful of short constructors.

**Tip:** make dynamic content (headers, choice lists) *functions* rather than values, so a screen you come back to after a `Pop` shows fresh data.

### 5. Data-driven multi-step flows (wizards)

When the steps of a flow depend on a type, describe the steps as **data** rather than code:

```go
kinds := []Kind{
    {Name: "A", Steps: []Step{title, date, priority}},
    {Name: "B", Steps: []Step{title, date}},
}
```

A generic `Wizard` screen walks the list. Each step can carry a `Skip func(draft) bool` for conditional questions. Adding a new kind means adding one table entry. The wizard itself never changes.

### Supporting principles

- **Separate domain state from UI state.** The lists themselves live in a `Store`. Screens get a pointer to whatever part of it they need. UI-only state, like cursors, lives in the screen. This also prepares you for the `todo-web` command, which can reuse the same domain and store with a different front end.
- **Pass context explicitly.** Don't keep "currently selected list" in a global. The entries screen is *constructed with* the `*List` it shows.
- **Keep keys local.** Each primitive decides what `b`, `esc`, and `enter` mean for itself. For example, a text prompt must not treat `b` as "back", or you can't type the letter b.

### Trade-offs

- **Closures vs. messages.** The callbacks below mutate the store directly, which is simple and fine for a local app. A stricter Elm style would have callbacks emit messages, like `addListMsg{name}`, that a single place applies. That's easier to test and log, but more boilerplate.
- **Stack vs. flat router.** A stack fits "drill-down" apps like this one. Tabbed apps fit better with a map of screens and a "current tab" key. The `Screen` interface works for both.
- **Libraries.** Charm's `huh` library provides ready-made forms and wizards for Bubble Tea. Writing your own (below) is worth doing once to learn the pattern. For a bigger project, `huh` could replace the Prompt and Wizard code.

---

# Part 2: Applying it to Lists4u

## New file layout

```
configs.go, todo_model.go,      ← removed
todo_logic.go, key_bindings.go

nav.go          Screen interface + Push/Pop/Replace commands
app.go          App: the one tea.Model, owns the screen stack
store.go        Store: the lists (domain state, no UI)
menu.go         Menu + NewConfirm (reusable)
prompt.go       Prompt (reusable)
wizard.go       Draft, Field, Wizard (reusable)
entry_kinds.go  The table of entry types and their fields
screens.go      The app's actual screens, each a small constructor

list.go, todo.go, event.go, ...  unchanged except one helper on List
cmd/todo/main.go                 todo.NewModel() → todo.NewApp()
```

## Old screen → new design

| Old `Screen` constant | Becomes |
|---|---|
| `HomeScreen` | `NewHomeScreen`: a root `Menu` |
| `AddListScreen` | `NewCreateListScreen`: a `Prompt` |
| `ViewListsScreen`, `RemoveListScreen` | `NewViewListsScreen` / `NewDeleteListsScreen`: both use one shared `newListPicker` |
| `ListRemovalVerificationScreen`, `EntryRemovalVerificationScreen` | `NewConfirm(...)` |
| `ViewListEntriesScreen` | `NewListEntriesScreen(list)` |
| `RemoveEntryScreen` | `NewRemoveEntryScreen(list)` |
| `AddEntryScreen` | `NewEntryTypeScreen(list)`, built from the `EntryKinds` table |
| `AddTitle/Description/DueDate/Priority/Online/LocationScreen` | Six `Field`s run by one `Wizard` |
| `listPendingRemoval`, `entryPendingRemoval`, `currentlySelectedList`, `currentEntry`, `inputPrefix`, `headerMessage`, `choices` | Gone. Each is captured in a closure or owned by a screen. |

## The code

### `nav.go`: the contract and navigation

```go
package todo

import tea "charm.land/bubbletea/v2"

// Screen is one self-contained page of the UI. It owns its own state,
// handles its own keys, and renders itself.
type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
}

// Navigation is requested with messages, so screens never need to know
// about the stack or about each other's internals.
type pushMsg struct{ screen Screen }
type replaceMsg struct{ screen Screen }
type popMsg struct{ n int }
type popToRootMsg struct{}

func Push(s Screen) tea.Cmd    { return func() tea.Msg { return pushMsg{s} } }
func Replace(s Screen) tea.Cmd { return func() tea.Msg { return replaceMsg{s} } }
func Pop() tea.Cmd             { return PopN(1) }
func PopN(n int) tea.Cmd       { return func() tea.Msg { return popMsg{n} } }
func PopToRoot() tea.Cmd       { return func() tea.Msg { return popToRootMsg{} } }
```

### `app.go`: the root model

```go
package todo

import tea "charm.land/bubbletea/v2"

// App is the only tea.Model. It holds a stack of screens and forwards
// every message to whichever screen is on top.
type App struct {
	stack []Screen
}

func NewApp() *App {
	return &App{stack: []Screen{NewHomeScreen(&Store{})}}
}

func (a *App) top() Screen { return a.stack[len(a.stack)-1] }

func (a *App) Init() tea.Cmd { return a.top().Init() }

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

func (a *App) View() tea.View { return tea.NewView(a.top().View()) }
```

This is the *entire* routing layer. Compare it with `keyLogic` and `determineSelection`.

### `store.go`: domain state, no UI

```go
package todo

import "slices"

// Store is the app's data, kept separate from any UI state.
type Store struct {
	Lists []*List
}

func (s *Store) AddList(name string) { s.Lists = append(s.Lists, NewList(name)) }

func (s *Store) RemoveList(i int) { s.Lists = slices.Delete(s.Lists, i, i+1) }
```

Using `[]*List` instead of `[]List` fixes the `go vet` "copies lock" warnings. It also means a screen holding a `*List` always sees that list's latest state.

Add this helper to `list.go`, so the UI stops reaching into `activeEntries` directly:

```go
// Active returns the entries that have not been removed, in insertion order.
func (l *List) Active() []Entry {
	entries := make([]Entry, 0, len(l.activeEntries))
	for _, e := range l.Listables {
		if _, ok := l.activeEntries[e.id]; ok {
			entries = append(entries, e.Entry)
		}
	}
	return entries
}
```

### `menu.go`: reusable menu and confirm

```go
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

// Menu is a reusable "header + list of choices" screen. Header and Choices
// are functions so they are re-read on every render and never go stale.
type Menu struct {
	Header  func() string
	Choices func() []Choice
	IsRoot  bool // the root menu has no "back"

	cursor int
}

// Text and Static adapt fixed values to Menu's function fields.
func Text(s string) func() string          { return func() string { return s } }
func Static(cs ...Choice) func() []Choice { return func() []Choice { return cs } }

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
		}
	case "down", "j":
		if m.cursor < len(choices)-1 {
			m.cursor++
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
```

The `len(choices) > 0` guard is what fixes the "View Lists with no lists" panic, and it fixes it for *every* menu at once.

### `prompt.go`: reusable text input

```go
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
```

Each prompt owns its own `textinput.Model`. You no longer have to `Blur()`, `SetValue("")`, and swap the placeholder on a shared input between screens. A new prompt starts clean.

> Behavior change: "go back" from a text prompt is now `esc` instead of `ctrl+c`. `ctrl+c` now always quits. `b` can't be used here because the user needs to type it.

### `wizard.go`: generic multi-step flow

```go
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
```

The wizard is a screen that *contains* another screen. This is the composite part of the pattern. Each wizard starts with a fresh `Draft`, so stale data can't leak from one entry into the next. When the child calls `Pop()` (user pressed `b`/`esc`), it pops the whole wizard, which cancels the entry.

### `entry_kinds.go`: the entry types as data

This replaces all of `determineNextField`, the `AddEntryScreen` cursor switch, and the per-type prompts in `UpdateChoices`.

```go
package todo

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// EntryKind describes everything the UI needs to create one type of entry.
// Adding a new entry type means adding one element to EntryKinds.
type EntryKind struct {
	Type   EntryType
	Fields []Field
	Build  func(Draft) Entry
}

var EntryKinds = []EntryKind{
	{
		Type:   todoType,
		Fields: []Field{titleField, descriptionField, whenField("When is this due?"), priorityField},
		Build: func(d Draft) Entry {
			return NewToDo(d.Title, d.Description, d.When, int(d.Priority))
		},
	},
	{
		Type:   reminderType,
		Fields: []Field{titleField, descriptionField, whenField("When do you want to be reminded?")},
		Build: func(d Draft) Entry {
			return NewReminder(d.Title, d.Description, d.When)
		},
	},
	{
		Type:   meetingType,
		Fields: []Field{titleField, descriptionField, whenField("When does this take place?"), priorityField, onlineField, locationField},
		Build: func(d Draft) Entry {
			return NewMeeting(d.Title, d.Description, d.Location, d.Online, d.When, d.Priority)
		},
	},
	{
		Type:   eventType,
		Fields: []Field{titleField, descriptionField, whenField("When does this take place?"), priorityField, locationField},
		Build: func(d Draft) Entry {
			return NewEvent(d.Title, d.Description, d.Location, d.When, d.Priority)
		},
	},
}

var titleField = Field{Screen: func(d *Draft) Screen {
	return NewPrompt("Name of Entry:", "Add Title", func(v string) tea.Cmd {
		d.Title = v
		return nextField
	})
}}

var descriptionField = Field{Screen: func(d *Draft) Screen {
	return NewPrompt("Description of Entry:", "Add Description", func(v string) tea.Cmd {
		d.Description = v
		return nextField
	})
}}

func whenField(question string) Field {
	return Field{Screen: func(d *Draft) Screen {
		p := NewPrompt(question+" (example 2006-01-02 15:04:05)", "Add DateTime", func(v string) tea.Cmd {
			d.When, _ = time.Parse(time.DateTime, v) // already validated
			return nextField
		})
		p.Validate = func(v string) error {
			_, err := time.Parse(time.DateTime, v)
			return err
		}
		return p
	}}
}

var priorityField = Field{Screen: func(d *Draft) Screen {
	options := []struct {
		label string
		value Priority
	}{{"light", light}, {"medium", medium}, {"high", high}, {"urgent", Urgent}}

	choices := make([]Choice, len(options))
	for i, o := range options {
		choices[i] = Choice{o.label, func() tea.Cmd {
			d.Priority = o.value
			return nextField
		}}
	}
	return &Menu{Header: Text("What priority does this entry have?"), Choices: Static(choices...)}
}}

var onlineField = Field{Screen: func(d *Draft) Screen {
	answer := func(online bool) func() tea.Cmd {
		return func() tea.Cmd {
			d.Online = online
			return nextField
		}
	}
	return &Menu{
		Header:  Text("Is this an online meeting?"),
		Choices: Static(Choice{"yes", answer(true)}, Choice{"no", answer(false)}),
	}
}}

var locationField = Field{
	Skip: func(d *Draft) bool { return d.Online },
	Screen: func(d *Draft) Screen {
		return NewPrompt("What location does this entry have?", "Add Location", func(v string) tea.Cmd {
			d.Location = v
			return nextField
		})
	},
}
```

What this table gives you:

- **The priority choice now works.** Each option sets its own value.
- **Dates are validated.** A bad date shows the parse error and keeps the user on that question.
- **Conditional steps are explicit.** "Skip location if online" is one line instead of a cursor check inside a switch.
- **New entry types are one table entry.** To add, say, a `Deadline`, write the domain type in `deadline.go` and add one element to `EntryKinds`. No other UI file changes.

### `screens.go`: the actual app

Each screen is now a few lines of wiring.

```go
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
```

Things to notice:

- **Back works without special cases.** From the entries screen, `b` returns to the list picker because that's what is under it on the stack. If you later open an entries screen from somewhere else, `b` will return *there* with no changes.
- **The "pending removal" fields are gone.** The index `i` is captured by the closure that needs it. (Go 1.22+ gives each loop iteration its own `i`, so this is safe.)
- **View and delete share one picker.** They differ only in the header and the action.

### `cmd/todo/main.go`

```go
p := tea.NewProgram(todo.NewApp())
```

## Migration steps

You can do this incrementally rather than all at once:

1. Add `store.go` and `List.Active()`. Switch `Model.lists` to `[]*List`. The `go vet` warnings go away.
2. Add `nav.go`, `app.go`, `menu.go`, and `prompt.go`. Port the **Home**, **Create List**, **View Lists**, and **Delete Lists** screens. To keep the not-yet-ported entry screens working in the meantime, temporarily wrap the old `Model` as a `Screen` adapter that you `Push` from View Lists.
3. Port the entries screen and remove-entry screen.
4. Add `wizard.go` and `entry_kinds.go`. Port entry creation.
5. Delete `configs.go`, `todo_model.go`, `todo_logic.go`, and `key_bindings.go`.
6. Write tests. Screens are now plain structs, so you can test one in isolation: build it, call `Update` with a `tea.KeyPressMsg`, and assert on `View()` or on the returned command's message.

## Where to go next

- **Split packages.** Move the UI files into a `tui/` package, leaving the domain (`list.go`, entry types, `store.go`) in the root. You'll need to export a few names (`todoType` → `TodoType`, `light` → `Light`, etc.). The payoff is that `cmd/todo-web` can import the domain and store without pulling in Bubble Tea.
- **Key bindings and help.** Use `charm.land/bubbles/v2/key` (`key.NewBinding`, `key.Matches`) instead of raw strings. Have each screen expose a `KeyMap`, so the footer can be generated by the `help` bubble instead of hand-written per screen.
- **Messages instead of closures.** If the store grows (saving to disk, syncing with the web app), have callbacks return messages like `addEntryMsg{list, entry}` and apply them in one place. That place becomes the natural spot for persistence and error handling.
- **Styling.** With rendering isolated in `Menu.View` and `Prompt.View`, adding `lipgloss` styles in those two places restyles the whole app.
