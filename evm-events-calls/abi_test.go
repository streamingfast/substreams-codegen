package evm_events_calls

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamesThatNeedDoublePluralization(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]bool
	}{
		{
			name:     "empty input returns empty map",
			input:    []string{},
			expected: map[string]bool{},
		},
		{
			name:     "single name with no conflicts",
			input:    []string{"Transfer"},
			expected: map[string]bool{},
		},
		{
			name:     "names that pluralize differently don't conflict",
			input:    []string{"Transfer", "Approval"},
			expected: map[string]bool{},
		},
		{
			name:  "name that conflicts with plural form of another",
			input: []string{"Transfer", "Transfers"},
			expected: map[string]bool{
				"Transfers": true,
			},
		},
		{
			name:  "multiple names with one conflict",
			input: []string{"Status", "Statuses", "Transfer"},
			expected: map[string]bool{
				"Statuses": true,
			},
		},
		{
			name:  "case with underscores - singular and plural collision",
			input: []string{"user_status", "user_statuses"},
			expected: map[string]bool{
				"user_statuses": true,
			},
		},
		{
			name:  "camelCase names with collision",
			input: []string{"UserStatus", "UserStatuses"},
			expected: map[string]bool{
				"UserStatuses": true,
			},
		},
		{
			name:  "multiple collisions",
			input: []string{"Status", "Statuses", "Address", "Addresses"},
			expected: map[string]bool{
				"Statuses":  true,
				"Addresses": true,
			},
		},
		{
			name:  "names with special pluralization rules",
			input: []string{"Child", "Children", "Person", "People"},
			expected: map[string]bool{
				"Children": true,
				"People":   true,
			},
		},
		{
			name:     "names that don't change when pluralized",
			input:    []string{"Data", "Series", "Species"},
			expected: map[string]bool{},
		},
		{
			name:  "mix of colliding and non-colliding names",
			input: []string{"Transfer", "Transfers", "Approve", "Mint", "Burn"},
			expected: map[string]bool{
				"Transfers": true,
			},
		},
		{
			name:  "names with numbers",
			input: []string{"ERC20", "ERC20s", "ERC721"},
			expected: map[string]bool{
				"ERC20s": true,
			},
		},
		{
			name:  "names with underscores at start (sanitization test)",
			input: []string{"_Transfer", "_Transfers"},
			expected: map[string]bool{
				"_Transfers": true,
			},
		},
		{
			name:  "names with multiple underscores (sanitization test)",
			input: []string{"foo___bar", "foo___bars"},
			expected: map[string]bool{
				"foo___bars": true,
			},
		},
		{
			name: "complex scenario with multiple types of names",
			input: []string{
				"Transfer",
				"Transfers",
				"Approval",
				"Status",
				"Statuses",
				"Update",
				"Data",
			},
			expected: map[string]bool{
				"Transfers": true,
				"Statuses":  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := namesThatNeedDoublePluralization(tt.input)

			// Check that all expected keys are present with true value
			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				assert.True(t, exists, "Expected key '%s' to be present in result", key)
				assert.Equal(t, expectedValue, actualValue, "Expected value for key '%s' to be %v", key, expectedValue)
			}

			// Check that no unexpected keys are present
			for key := range result {
				_, expected := tt.expected[key]
				assert.True(t, expected, "Unexpected key '%s' in result", key)
			}

			// Alternatively, can use require.Equal for exact map comparison
			require.Equal(t, tt.expected, result, "Maps should be equal")
		})
	}
}
