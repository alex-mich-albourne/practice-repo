package todo

import (
	"fmt"
	"time"
)

type Reminder struct {
	title        string
	description  string
	reminderTime time.Time
}

func (r *Reminder) Type() EntryType {
	return reminderType
}

func (r *Reminder) String() string {
	return fmt.Sprintf("%s %s | %v\nDescription: %s", r.Type(), r.title, r.reminderTime, r.description)
}

func (r *Reminder) Priority() Priority {
	return none
}

func NewReminder(title, description string, reminderTime time.Time) *Reminder {
	return &Reminder{
		title:        title,
		description:  description,
		reminderTime: reminderTime,
	}
}
