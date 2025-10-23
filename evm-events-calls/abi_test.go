package evm_events_calls

import (
	"testing"

	"github.com/streamingfast/eth-go"
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

func TestIndexedDynamicValueHandling(t *testing.T) {
	// Test ABI with indexed string parameter (dynamic type)
	abiJSON := `[
		{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"internalType": "string",
					"name": "param0",
					"type": "string"
				},
				{
					"indexed": false,
					"internalType": "uint256",
					"name": "value",
					"type": "uint256"
				}
			],
			"name": "EventStringIdx",
			"type": "event"
		}
	]`

	abi, err := eth.ParseABIFromBytes([]byte(abiJSON))
	require.NoError(t, err)

	abiWrapper := &ABI{abi: abi, raw: abiJSON}
	events, err := abiWrapper.BuildEventModels()
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	
	// Check that the indexed dynamic parameter is handled correctly
	// The transformation code should handle IndexedDynamicValue<String>
	transformCode, exists := event.Rust.ProtoFieldABIConversionMap["param0"]
	require.True(t, exists, "param0 should have transformation code")
	
	// For indexed dynamic values, we expect the transformation to access the hash field
	assert.Contains(t, transformCode, ".hash", "Indexed dynamic string should use .hash field")
	
	// The non-indexed parameter should use normal transformation
	transformCode2, exists := event.Rust.ProtoFieldABIConversionMap["value"]
	require.True(t, exists, "value should have transformation code")
	assert.NotContains(t, transformCode2, ".hash", "Non-indexed parameter should not use .hash field")
	
	// Check proto field types
	require.Len(t, event.Proto.Fields, 2, "Should have 2 proto fields")
	
	// Find the param0 field (indexed dynamic string)
	var param0Field, valueField *ProtoField
	for i := range event.Proto.Fields {
		if event.Proto.Fields[i].Name == "param0" {
			param0Field = &event.Proto.Fields[i]
		} else if event.Proto.Fields[i].Name == "value" {
			valueField = &event.Proto.Fields[i]
		}
	}
	
	require.NotNil(t, param0Field, "param0 field should exist")
	require.NotNil(t, valueField, "value field should exist")
	
	// Indexed dynamic string should be bytes (hash)
	assert.Equal(t, "bytes", param0Field.Type, "Indexed dynamic string should be bytes proto field")
	
	// Non-indexed uint256 should be string
	assert.Equal(t, "string", valueField.Type, "uint256 should be string proto field")
}

func TestIsDynamicType(t *testing.T) {
	tests := []struct {
		name     string
		abiJSON  string
		expected bool
	}{
		{
			name: "string type is dynamic",
			abiJSON: `[{
				"inputs": [{"type": "string", "name": "test"}],
				"name": "TestEvent",
				"type": "event"
			}]`,
			expected: true,
		},
		{
			name: "bytes type is dynamic",
			abiJSON: `[{
				"inputs": [{"type": "bytes", "name": "test"}],
				"name": "TestEvent",
				"type": "event"
			}]`,
			expected: true,
		},
		{
			name: "uint256 type is not dynamic",
			abiJSON: `[{
				"inputs": [{"type": "uint256", "name": "test"}],
				"name": "TestEvent",
				"type": "event"
			}]`,
			expected: false,
		},
		{
			name: "address type is not dynamic",
			abiJSON: `[{
				"inputs": [{"type": "address", "name": "test"}],
				"name": "TestEvent",
				"type": "event"
			}]`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abi, err := eth.ParseABIFromBytes([]byte(tt.abiJSON))
			require.NoError(t, err)
			
			events := abi.LogEventsByNameMap["TestEvent"]
			require.Len(t, events, 1)
			
			param := events[0].Parameters[0]
			result := isDynamicType(param.Type)
			assert.Equal(t, tt.expected, result)
		})
	}
}
