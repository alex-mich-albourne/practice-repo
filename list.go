package todo

import (
	"errors"
	"strings"
	"sync"
)

type List struct {
	Name          string
	Listables     []Listable
	LastEntry     int
	activeEntries map[int]*Listable

	mu sync.Mutex
}
type Listable struct {
	Entry

	id int
}

type Entry interface {
	String() string
	Type() EntryType
	Priority() Priority
}

type Priority string

const (
	none   Priority = "none"
	light  Priority = "light"
	medium Priority = "medium"
	high   Priority = "high"
	Urgent Priority = "urgent"
)

type EntryType string

const (
	todoType     EntryType = "ToDo"
	reminderType EntryType = "Reminder"
	eventType    EntryType = "Event"
	meetingType  EntryType = "Meeting"
	ticketType   EntryType = "Ticket"
)

func NewList(name string) *List {
	return &List{
		Name:          name,
		Listables:     make([]Listable, 0),
		LastEntry:     -1,
		activeEntries: make(map[int]*Listable),
	}
}

func NewListable(e Entry, id int) *Listable {
	return &Listable{
		Entry: e,
		id:    id,
	}
}

func (t *List) String() string {
	sb := strings.Builder{}
	length := len(t.activeEntries)
	i := 0
	for _, entry := range t.Listables {
		_, ok := t.activeEntries[entry.id]
		if !ok {
			continue
		}
		sb.WriteString(entry.String())
		if i != length-1 {
			sb.WriteString("\n\n")
		}
		i++
	}
	return sb.String()
}

func (l *List) AddEntry(e Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.LastEntry++
	id := l.LastEntry
	newListable := NewListable(e, id)
	l.Listables = append(l.Listables, *newListable)
	l.activeEntries[id] = newListable
}

var ErrNotFound = errors.New("not found")

func (l *List) RemoveEntry(id int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.activeEntries[id]
	if ok {
		delete(l.activeEntries, id)
		return nil
	}
	return ErrNotFound
}

func (l *List) GetEntry(eId int) Entry {
	return l.Listables[eId]
}

func (l *List) GetAllEntries() []Entry {
	entries := make([]Entry, 0, len(l.Listables))
	for _, e := range l.Listables {
		entries = append(entries, e)
	}
	return entries
}

func (l *List) GetEntriesByPriority(priority Priority) []Entry {
	entries := make([]Entry, 0, len(l.Listables))
	for _, e := range l.Listables {
		if e.Priority() == priority {
			entries = append(entries, e)
		}
	}
	return entries
}

func (l *List) DeactivateListable(pos int) {
	i := 0
	for _, listable := range l.Listables {
		_, ok := l.activeEntries[listable.id]
		if !ok {
			continue
		}
		if i == pos {
			if err := l.RemoveEntry(listable.id); err != nil {
				panic(err)
			}
			break
		}
		i++
	}
}

func (l *List) Active() []Entry {
	entries := make([]Entry, 0, len(l.activeEntries))
	for _, e := range l.Listables {
		if _, ok := l.activeEntries[e.id]; ok {
			entries = append(entries, e.Entry)
		}
	}
	return entries
}
