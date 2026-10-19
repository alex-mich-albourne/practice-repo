package todo

import "testing"

func Test_List_String(t *testing.T) {
	var tests = []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "list with entries",
			input:    []string{"todo 1", "todo 2", "todo 3", "todo 4"},
			expected: "todo 1\n\ntodo 2\n\ntodo 3\n\ntodo 4",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			list := NewList(test.name)
			for _, name := range test.input {
				list.AddEntry(mockEntry{name: name})
			}

			result := list.String()

			if result != test.expected {
				t.Errorf("expected: %s, got: %s", test.expected, result)
			}
		})
	}
}

func Test_List_DeactivateListable(t *testing.T) {
	var tests = []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "list with entries",
			input:    []string{"todo 1", "todo 2", "todo 3", "todo 4"},
			expected: "todo 1\n\ntodo 3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			list := NewList(test.name)
			for _, name := range test.input {
				list.AddEntry(mockEntry{name: name})
			}

			list.DeactivateListable(1)
			list.DeactivateListable(2)

			result := list.String()

			if result != test.expected {
				t.Errorf("expected: %s, got: %s", test.expected, result)
			}
		})
	}
}

type mockEntry struct {
	name string
}

func (m mockEntry) Type() EntryType {
	return todoType
}
func (m mockEntry) Priority() Priority {
	return Priority(0)
}

func (m mockEntry) String() string {
	return m.name
}
