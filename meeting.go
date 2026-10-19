package todo

import (
	"fmt"
	"time"
)

type Meeting struct {
	title       string
	description string
	meetingTime time.Time
	online      bool
	location    string
	priority    Priority
}

func (r *Meeting) Type() EntryType {
	return meetingType
}

func (r *Meeting) String() string {
	var locationString string = ""
	if r.online {
		locationString = "| online"
	} else {
		locationString = fmt.Sprintf("at %s", r.location)
	}
	return fmt.Sprintf(
		"%s for %s %s | Date: %v | Time: %v\nDescription: %s",
		r.Type(),
		r.title,
		locationString,
		r.meetingTime.Format("2006-01-02"),
		r.meetingTime.Format(time.TimeOnly),
		r.description)
}

func (r *Meeting) Priority() Priority {
	return r.priority
}

func NewMeeting(title, description, location string, online bool, meetingTime time.Time, priority Priority) *Meeting {
	return &Meeting{
		title:       title,
		description: description,
		meetingTime: meetingTime,
		online:      online,
		location:    location,
		priority:    priority,
	}
}
