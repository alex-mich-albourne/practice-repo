package todo

import (
	"fmt"
	"time"
)

type ToDo struct {
	title       string
	description string
	dueDate     time.Time
	priority    Priority
}

func (t *ToDo) Type() EntryType {
	return todoType
}

func (t *ToDo) String() string {
	return fmt.Sprintf("%s | Type: %s\nDue on: %v\nDescription: %s",
		t.title, t.Type(), t.dueDate.Format("2006-01-02"), t.description)
}

func (t *ToDo) Priority() Priority {
	return t.priority
}

func NewToDo(title, description string, dueDate time.Time, priority Priority) *ToDo {
	return &ToDo{
		title:       title,
		description: description,
		dueDate:     dueDate,
		priority:    priority,
	}
}
