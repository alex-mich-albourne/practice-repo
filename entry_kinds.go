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
			return NewToDo(d.Title, d.Description, d.When, d.Priority)
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
	{
		Type:   ticketType,
		Fields: []Field{titleField, descriptionField, whenField("When is this ticket due?"), priorityField, platformField},
		Build: func(d Draft) Entry {
			return NewTicket(d.Title, d.Description, d.Platform, d.When, d.Priority)
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

var platformField = Field{Screen: func(d *Draft) Screen {
	return NewPrompt("What Platform is the issue on?", "Enter Platform", func(v string) tea.Cmd {
		d.Platform = v
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
	}{
		{"light", light},
		{"medium", medium},
		{"high", high},
		{"urgent", Urgent},
	}

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
