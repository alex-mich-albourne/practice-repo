package todo

import (
	"fmt"
	"time"
)

type Ticket struct {
	title       string
	description string
	due         time.Time
	priority    Priority
	platform    string
}

func (t *Ticket) Type() EntryType { return ticketType }

func (t *Ticket) String() string {
	return fmt.Sprintf(
		"%s: %s | %s Priority | Due on: %v\nDescription: %s on the %s platform",
		t.Type(),
		t.title,
		t.priority,
		t.due,
		t.description,
		t.platform,
	)
}

func (t *Ticket) Priority() Priority {
	return t.priority
}

func NewTicket(title, description, platform string, when time.Time, priority Priority) *Ticket {
	return &Ticket{
		title:       title,
		description: description,
		due:         when,
		priority:    priority,
		platform:    platform,
	}
}
