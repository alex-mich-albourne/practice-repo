package todo

import (
	"fmt"
	"time"
)

type Event struct {
	title       string
	description string
	eventTime   time.Time
	location    string
	priority    Priority
}

func (r *Event) Type() EntryType {
	return eventType
}

func (r *Event) String() string {
	return fmt.Sprintf("%s: %s | %v | Location: %s\nDescription: %s", r.Type(), r.title, r.eventTime, r.location, r.description)
}

func (r *Event) Priority() Priority {
	return r.priority
}

func NewEvent(title, description, location string, eventTime time.Time, priority Priority) *Event {
	return &Event{
		title:       title,
		description: description,
		eventTime:   eventTime,
		location:    location,
		priority:    priority,
	}
}
