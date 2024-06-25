package injective_events

import (
	"testing"
)

func TestGetEventsQuery(t *testing.T) {
	testCases := []struct {
		name       string
		eventDescs []*eventDesc
		expected   string
	}{
		// Test case 1: No event descriptions
		{
			name:       "No event descriptions",
			eventDescs: []*eventDesc{},
			expected:   "",
		},
		// Test case 2: Single event description with no attributes
		{
			name: "Single event description with no attributes",
			eventDescs: []*eventDesc{
				{
					EventType:  "event1",
					Attributes: nil,
				},
			},
			expected: "type:event1",
		},
		// Test case 3: Single event description with attributes
		{
			name: "Single event description with attributes without value",
			eventDescs: []*eventDesc{
				{
					EventType: "Event2",
					Attributes: map[string]string{
						"attr1": "",
						"attr2": "usdt",
					},
				},
			},
			expected: "(type:Event2 && (attr:attr1 && attr:attr2:usdt))",
		},
		// Test case 4: Multiple event descriptions
		{
			name: "Multiple event descriptions",
			eventDescs: []*eventDesc{
				{
					EventType: "Event3",
					Attributes: map[string]string{
						"attr1": "",
					},
				},
				{
					EventType: "Event4",
					Attributes: map[string]string{
						"attr2": "usdt",
						"attr3": "inj",
					},
				},
			},
			expected: "(type:Event3 && (attr:attr1)) || (type:Event4 && (attr:attr2:usdt && attr:attr3:inj))",
		},
		// Additional test cases can be added here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			project := &Project{
				EventDescs: tc.eventDescs,
			}
			if got := project.GetEventsQuery(); got != tc.expected {
				t.Errorf("Test case %s failed: expected %q, got %q", tc.name, tc.expected, got)
			}
		})
	}
}
