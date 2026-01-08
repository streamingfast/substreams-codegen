# Implementation Plan for Codegen Agent

## Objective

Implement support for wrapped ABI formats (Hardhat/Foundry/Truffle artifacts) in the `substreams-codegen` EVM generators. The implementation should be backward compatible, accepting both the standard ABI array format and wrapped object format with an `abi` field.

## How to Use This Plan

- Follow steps sequentially, completing each before moving to the next
- Verify your work at each checkpoint using the specified commands
- If you encounter ambiguity, refer to the linked source files for context
- Run all verification commands from the project root directory
- Ensure all tests pass before considering a step complete
- Create a pull request when implementation is complete

---

## Requirements Summary

### Current Behavior

Only accepts direct ABI array format:
```json
[
  { "type": "function", "name": "transfer", ... },
  { "type": "event", "name": "Transfer", ... }
]
```

### Required Behavior

1. First try to parse as standard ABI array format (current behavior)
2. If that fails, check if the JSON is an object with an "abi" field
3. If yes, extract the ABI from the "abi" field and use it
4. Ignore all other fields in the wrapper object (contractName, bytecode, etc.)
5. Provide clear error messages if neither format works

### Wrapped Format Example (Hardhat/Foundry)

```json
{
   "contractName": "MyContract",
   "abi": [ /* actual ABI array here */ ],
   "bytecode": "0x...",
   "deployedBytecode": "0x...",
   "other_fields": "..."
}
```

### Affected Input Methods

- API fetch from Etherscan (already handles this via `gjson`)
- Manual JSON string input in conversation
- Local file path input

---

## Implementation Steps

### Step 1: Create ABI Extraction Helper Function

**File to modify:** `./evm-events-calls/abi.go`

Add a new helper function that handles both formats. Insert this after the existing imports and before `CmdDecodeABI`:

```go
// extractABIFromJSON attempts to extract a valid ABI JSON array from the input.
// It handles two formats:
// 1. Standard ABI array format: [{"type": "function", ...}, ...]
// 2. Wrapped format (Hardhat/Foundry): {"abi": [...], "contractName": "...", ...}
//
// Returns the raw ABI JSON bytes ready for parsing, or an error if neither format is valid.
func extractABIFromJSON(input []byte) ([]byte, error) {
	// Trim whitespace for consistent detection
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty ABI input")
	}

	// Check if it starts with '[' - standard ABI array format
	if trimmed[0] == '[' {
		// Validate it's proper JSON
		if !json.Valid(trimmed) {
			return nil, fmt.Errorf("invalid JSON array format")
		}
		return trimmed, nil
	}

	// Check if it starts with '{' - potentially wrapped format
	if trimmed[0] == '{' {
		// Try to extract the "abi" field from the object
		var wrapper struct {
			ABI json.RawMessage `json:"abi"`
		}
		if err := json.Unmarshal(trimmed, &wrapper); err != nil {
			return nil, fmt.Errorf("invalid JSON object format: %w", err)
		}

		// Check if the "abi" field exists and is not empty
		if len(wrapper.ABI) == 0 {
			return nil, fmt.Errorf("JSON object does not contain an 'abi' field or it is empty; expected either a JSON array (standard ABI format) or an object with an 'abi' field (Hardhat/Foundry artifact format)")
		}

		// Validate the extracted ABI is a valid JSON array
		abiTrimmed := bytes.TrimSpace(wrapper.ABI)
		if len(abiTrimmed) == 0 || abiTrimmed[0] != '[' {
			return nil, fmt.Errorf("the 'abi' field must be a JSON array, got: %s", string(abiTrimmed[:min(20, len(abiTrimmed))]))
		}

		return wrapper.ABI, nil
	}

	return nil, fmt.Errorf("invalid ABI format: expected JSON array or object, got input starting with %q", string(trimmed[:min(1, len(trimmed))]))
}
```

Also add the `bytes` import at the top of the file:

```go
import (
	"bytes"         // ADD THIS
	"encoding/hex"
	"encoding/json" // ADD THIS
	"fmt"
	// ... rest of imports
)
```

**Note:** The `encoding/json` import may already exist. Verify before adding.

### Step 2: Update CmdDecodeABI Function

**File:** `./evm-events-calls/abi.go`

Modify `CmdDecodeABI` to use the new extraction function:

**Current code (lines 26-31):**
```go
func CmdDecodeABI(contract *Contract) loop.Cmd {
	return func() loop.Msg {
		abi, err := eth.ParseABIFromBytes([]byte(contract.RawABI))
		return ReturnRunDecodeContractABI{abi: &ABI{abi, string(contract.RawABI)}, err: err}
	}
}
```

**New code:**
```go
func CmdDecodeABI(contract *Contract) loop.Cmd {
	return func() loop.Msg {
		abiBytes, err := extractABIFromJSON(contract.RawABI)
		if err != nil {
			return ReturnRunDecodeContractABI{abi: nil, err: fmt.Errorf("extracting ABI: %w", err)}
		}

		abi, err := eth.ParseABIFromBytes(abiBytes)
		if err != nil {
			return ReturnRunDecodeContractABI{abi: nil, err: fmt.Errorf("parsing ABI: %w", err)}
		}

		return ReturnRunDecodeContractABI{abi: &ABI{abi, string(abiBytes)}, err: nil}
	}
}
```

### Step 3: Update cmdDecodeDynamicABI Function

**File:** `./evm-events-calls/abi.go`

Apply the same change to the dynamic contract decoder:

**Current code (lines 33-38):**
```go
func cmdDecodeDynamicABI(contract *DynamicContract) loop.Cmd {
	return func() loop.Msg {
		abi, err := eth.ParseABIFromBytes([]byte(contract.RawABI))
		return ReturnRunDecodeDynamicContractABI{abi: &ABI{abi, string(contract.RawABI)}, err: err}
	}
}
```

**New code:**
```go
func cmdDecodeDynamicABI(contract *DynamicContract) loop.Cmd {
	return func() loop.Msg {
		abiBytes, err := extractABIFromJSON(contract.RawABI)
		if err != nil {
			return ReturnRunDecodeDynamicContractABI{abi: nil, err: fmt.Errorf("extracting ABI: %w", err)}
		}

		abi, err := eth.ParseABIFromBytes(abiBytes)
		if err != nil {
			return ReturnRunDecodeDynamicContractABI{abi: nil, err: fmt.Errorf("parsing ABI: %w", err)}
		}

		return ReturnRunDecodeDynamicContractABI{abi: &ABI{abi, string(abiBytes)}, err: nil}
	}
}
```

### Step 4: Update abiToJson Function in convo.go

**File:** `./evm-events-calls/convo.go`

The `abiToJson` function at the bottom of the file (lines 907-915) should also handle wrapped formats for early validation:

**Current code:**
```go
func abiToJson(abi string) (json.RawMessage, error) {
	var rawMessage json.RawMessage = json.RawMessage(abi)

	if _, err := json.Marshal(rawMessage); err != nil {
		return nil, err
	}

	return rawMessage, nil
}
```

**New code:**
```go
func abiToJson(input string) (json.RawMessage, error) {
	inputBytes := []byte(input)

	// Validate it's valid JSON first
	if !json.Valid(inputBytes) {
		return nil, fmt.Errorf("invalid JSON")
	}

	// Try to extract ABI (handles both array and wrapped formats)
	// This is just validation - the actual extraction happens in CmdDecodeABI
	trimmed := bytes.TrimSpace(inputBytes)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	// For array format, accept as-is
	if trimmed[0] == '[' {
		return json.RawMessage(inputBytes), nil
	}

	// For object format, validate it has an "abi" field
	if trimmed[0] == '{' {
		var wrapper struct {
			ABI json.RawMessage `json:"abi"`
		}
		if err := json.Unmarshal(trimmed, &wrapper); err != nil {
			return nil, fmt.Errorf("invalid JSON object: %w", err)
		}
		if len(wrapper.ABI) == 0 {
			return nil, fmt.Errorf("object has no 'abi' field; expected a JSON array or an object with an 'abi' field (Hardhat/Foundry format)")
		}
		// Return the full input - extraction happens later in CmdDecodeABI
		return json.RawMessage(inputBytes), nil
	}

	return nil, fmt.Errorf("expected JSON array or object")
}
```

Add the `bytes` import at the top of `convo.go`:

```go
import (
	"bytes"  // ADD THIS
	"encoding/json"
	"fmt"
	// ... rest
)
```

### Step 5: Add Unit Tests for ABI Extraction

**File:** `./evm-events-calls/abi_test.go`

Add comprehensive tests for the new `extractABIFromJSON` function:

```go
func TestExtractABIFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		wantABI     string // Expected extracted ABI (for non-error cases)
	}{
		{
			name:    "standard ABI array format",
			input:   `[{"type": "function", "name": "transfer"}]`,
			wantErr: false,
			wantABI: `[{"type": "function", "name": "transfer"}]`,
		},
		{
			name:    "standard ABI array format with whitespace",
			input:   `  [{"type": "event", "name": "Transfer"}]  `,
			wantErr: false,
			wantABI: `[{"type": "event", "name": "Transfer"}]`,
		},
		{
			name: "wrapped format - Hardhat style",
			input: `{
				"contractName": "MyContract",
				"abi": [{"type": "function", "name": "transfer"}],
				"bytecode": "0x1234"
			}`,
			wantErr: false,
			wantABI: `[{"type": "function", "name": "transfer"}]`,
		},
		{
			name: "wrapped format - Foundry style",
			input: `{
				"abi": [{"type": "event", "name": "Transfer"}],
				"bytecode": {"object": "0x..."},
				"methodIdentifiers": {}
			}`,
			wantErr: false,
			wantABI: `[{"type": "event", "name": "Transfer"}]`,
		},
		{
			name: "wrapped format - minimal",
			input: `{"abi": []}`,
			wantErr: false,
			wantABI: `[]`,
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: true,
			errContains: "empty ABI input",
		},
		{
			name:    "whitespace only",
			input:   `   `,
			wantErr: true,
			errContains: "empty ABI input",
		},
		{
			name:    "invalid JSON",
			input:   `{invalid json`,
			wantErr: true,
			errContains: "invalid JSON",
		},
		{
			name:    "object without abi field",
			input:   `{"contractName": "Test", "bytecode": "0x"}`,
			wantErr: true,
			errContains: "does not contain an 'abi' field",
		},
		{
			name:    "object with null abi",
			input:   `{"abi": null}`,
			wantErr: true,
			errContains: "does not contain an 'abi' field",
		},
		{
			name:    "object with non-array abi",
			input:   `{"abi": "not an array"}`,
			wantErr: true,
			errContains: "must be a JSON array",
		},
		{
			name:    "object with object abi",
			input:   `{"abi": {"key": "value"}}`,
			wantErr: true,
			errContains: "must be a JSON array",
		},
		{
			name:        "number input",
			input:       `123`,
			wantErr:     true,
			errContains: "expected JSON array or object",
		},
		{
			name:        "string input",
			input:       `"hello"`,
			wantErr:     true,
			errContains: "expected JSON array or object",
		},
		{
			name: "real Hardhat artifact format",
			input: `{
				"_format": "hh-sol-artifact-1",
				"contractName": "ERC20",
				"sourceName": "contracts/ERC20.sol",
				"abi": [
					{
						"inputs": [{"name": "to", "type": "address"}, {"name": "amount", "type": "uint256"}],
						"name": "transfer",
						"outputs": [{"name": "", "type": "bool"}],
						"stateMutability": "nonpayable",
						"type": "function"
					}
				],
				"bytecode": "0x608060...",
				"deployedBytecode": "0x608060...",
				"linkReferences": {},
				"deployedLinkReferences": {}
			}`,
			wantErr: false,
		},
		{
			name: "real Foundry artifact format",
			input: `{
				"abi": [
					{
						"type": "function",
						"name": "transfer",
						"inputs": [{"name": "to", "type": "address"}, {"name": "amount", "type": "uint256"}],
						"outputs": [{"name": "", "type": "bool"}],
						"stateMutability": "nonpayable"
					}
				],
				"bytecode": {"object": "0x608060...", "sourceMap": "..."},
				"deployedBytecode": {"object": "0x608060...", "sourceMap": "..."},
				"methodIdentifiers": {"transfer(address,uint256)": "a9059cbb"}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractABIFromJSON([]byte(tt.input))

			if tt.wantErr {
				require.Error(t, err, "expected an error but got none")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "error message should contain expected text")
				}
				return
			}

			require.NoError(t, err, "unexpected error")
			assert.NotEmpty(t, result, "result should not be empty")

			// Verify the result is valid JSON
			assert.True(t, json.Valid(result), "result should be valid JSON")

			// If expected ABI is specified, compare after normalization
			if tt.wantABI != "" {
				// Normalize both by unmarshaling and remarshaling
				var expected, actual interface{}
				require.NoError(t, json.Unmarshal([]byte(tt.wantABI), &expected))
				require.NoError(t, json.Unmarshal(result, &actual))
				assert.Equal(t, expected, actual, "extracted ABI should match expected")
			}
		})
	}
}
```

### Step 6: Add Integration Test for Wrapped ABI in Conversation Flow

**File:** `./evm-events-calls/convo_test.go`

Add a test case that verifies the wrapped ABI format works through the conversation flow:

```go
func TestWrappedABIFormat(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup: skip to the point where we need ABI input
	p.Name = "test-project"
	p.ChainName = "mainnet"
	p.Contracts = append(p.Contracts, &Contract{
		Address: "0x1231231230123123123012312312301231231230",
	})
	p.currentContractIdx = 0

	// Test with wrapped ABI format (Hardhat style)
	wrappedABI := `{
		"contractName": "TestContract",
		"abi": [
			{
				"anonymous": false,
				"inputs": [
					{"indexed": true, "name": "from", "type": "address"},
					{"indexed": true, "name": "to", "type": "address"},
					{"indexed": false, "name": "value", "type": "uint256"}
				],
				"name": "Transfer",
				"type": "event"
			}
		],
		"bytecode": "0x608060..."
	}`

	// Simulate file input with wrapped ABI
	next := conv.Update(InputContractABIFile{
		UserInput_LocalFile: pbconvo.UserInput_LocalFile{
			Value: []byte(wrappedABI),
		},
	})

	// Should proceed to RunDecodeContractABI
	assert.Equal(t, RunDecodeContractABI{}, next())

	// Execute the decode
	next = conv.Update(RunDecodeContractABI{})
	msg, ok := next().(ReturnRunDecodeContractABI)
	require.True(t, ok, "expected ReturnRunDecodeContractABI")
	require.NoError(t, msg.err, "should successfully decode wrapped ABI")
	require.NotNil(t, msg.abi, "ABI should not be nil")

	// Verify the ABI was correctly extracted
	events := p.Contracts[0].EventModels()
	require.Len(t, events, 1, "should have one event")
	assert.Equal(t, "Transfer", events[0].Proto.MessageName, "event name should be Transfer")
}

func TestWrappedABIFormatWithStringInput(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup
	p.Name = "test-project"
	p.ChainName = "mainnet"
	p.Contracts = append(p.Contracts, &Contract{
		Address: "0x1231231230123123123012312312301231231230",
	})
	p.currentContractIdx = 0
	p.Contracts[0].abiType = "string"

	// Test with wrapped ABI format via string input
	wrappedABI := `{"abi": [{"type": "function", "name": "transfer", "inputs": [], "outputs": [], "stateMutability": "nonpayable"}]}`

	next := conv.Update(InputContractABIString{
		UserInput_TextInput: pbconvo.UserInput_TextInput{
			Value: wrappedABI,
		},
	})

	// Should proceed to next step (not error out)
	assert.Equal(t, RunDecodeContractABI{}, next())
}

func TestInvalidWrappedABIFormat(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup
	p.Name = "test-project"
	p.ChainName = "mainnet"
	p.Contracts = append(p.Contracts, &Contract{
		Address: "0x1231231230123123123012312312301231231230",
	})
	p.currentContractIdx = 0

	// Test with object that has no "abi" field
	invalidWrapped := `{"contractName": "Test", "bytecode": "0x..."}`

	next := conv.Update(InputContractABIFile{
		UserInput_LocalFile: pbconvo.UserInput_LocalFile{
			Value: []byte(invalidWrapped),
		},
	})

	// Should show error and ask for ABI again
	seq := next().(loop.SeqMsg)
	msg := seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, msg.GetMessage().Markdown, "abi")
}
```

### Step 7: Add Test Data Files

**File to create:** `./evm-events-calls/testdata/wrapped_abi_hardhat.json`

```json
{
  "_format": "hh-sol-artifact-1",
  "contractName": "TestToken",
  "sourceName": "contracts/TestToken.sol",
  "abi": [
    {
      "anonymous": false,
      "inputs": [
        {"indexed": true, "internalType": "address", "name": "from", "type": "address"},
        {"indexed": true, "internalType": "address", "name": "to", "type": "address"},
        {"indexed": false, "internalType": "uint256", "name": "value", "type": "uint256"}
      ],
      "name": "Transfer",
      "type": "event"
    },
    {
      "inputs": [
        {"internalType": "address", "name": "to", "type": "address"},
        {"internalType": "uint256", "name": "amount", "type": "uint256"}
      ],
      "name": "transfer",
      "outputs": [{"internalType": "bool", "name": "", "type": "bool"}],
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ],
  "bytecode": "0x608060405234801561001057600080fd5b50",
  "deployedBytecode": "0x608060405234801561001057600080fd5b50",
  "linkReferences": {},
  "deployedLinkReferences": {}
}
```

**File to create:** `./evm-events-calls/testdata/wrapped_abi_foundry.json`

```json
{
  "abi": [
    {
      "type": "event",
      "name": "Transfer",
      "inputs": [
        {"name": "from", "type": "address", "indexed": true},
        {"name": "to", "type": "address", "indexed": true},
        {"name": "value", "type": "uint256", "indexed": false}
      ],
      "anonymous": false
    },
    {
      "type": "function",
      "name": "transfer",
      "inputs": [
        {"name": "to", "type": "address"},
        {"name": "amount", "type": "uint256"}
      ],
      "outputs": [{"name": "", "type": "bool"}],
      "stateMutability": "nonpayable"
    }
  ],
  "bytecode": {
    "object": "0x608060405234801561001057600080fd5b50",
    "sourceMap": "..."
  },
  "methodIdentifiers": {
    "transfer(address,uint256)": "a9059cbb"
  }
}
```

---

## File Modification Summary

| File | Changes |
|------|---------|
| `./evm-events-calls/abi.go` | Add `extractABIFromJSON()` function, update `CmdDecodeABI()` and `cmdDecodeDynamicABI()` |
| `./evm-events-calls/convo.go` | Update `abiToJson()` function, add `bytes` import |
| `./evm-events-calls/abi_test.go` | Add `TestExtractABIFromJSON` with comprehensive test cases |
| `./evm-events-calls/convo_test.go` | Add `TestWrappedABIFormat`, `TestWrappedABIFormatWithStringInput`, `TestInvalidWrappedABIFormat` |
| `./evm-events-calls/testdata/wrapped_abi_hardhat.json` | New file - Hardhat artifact test data |
| `./evm-events-calls/testdata/wrapped_abi_foundry.json` | New file - Foundry artifact test data |

---

## Verification Steps

### 1. Unit Tests

```bash
# Run all tests for the evm-events-calls package
go test ./evm-events-calls/... -v

# Run specific new tests
go test ./evm-events-calls/... -v -run TestExtractABIFromJSON
go test ./evm-events-calls/... -v -run TestWrappedABI
```

**Expected:** All tests pass, including new tests for wrapped ABI formats.

### 2. Build Verification

```bash
# Build all commands
go build ./cmd/...

# Verify no compilation errors
echo $?  # Should be 0
```

**Expected:** Clean build with no warnings or errors.

### 3. Backward Compatibility Check

```bash
# Run the existing test suite to ensure nothing is broken
go test ./evm-events-calls/... -v

# The existing tests in convo_test.go and abi_test.go should still pass
```

**Expected:** All existing tests continue to pass without modification.

---

## Error Messages Reference

The implementation should produce these user-friendly error messages:

| Scenario | Error Message |
|----------|---------------|
| Empty input | `empty ABI input` |
| Invalid JSON | `invalid JSON array format` or `invalid JSON object format: <details>` |
| Object without `abi` field | `JSON object does not contain an 'abi' field or it is empty; expected either a JSON array (standard ABI format) or an object with an 'abi' field (Hardhat/Foundry artifact format)` |
| `abi` field is not an array | `the 'abi' field must be a JSON array, got: <preview>` |
| Unexpected format | `invalid ABI format: expected JSON array or object, got input starting with "<char>"` |

---

## Project Standards Checklist

- [ ] Code follows Go conventions and existing patterns in the repository
- [ ] Error messages are descriptive and actionable
- [ ] New functions have appropriate documentation comments
- [ ] Table-driven tests using testify assert/require
- [ ] Uses `require.NoError(t, err)` for all error checking in tests
- [ ] Test helpers call `t.Helper()` as first statement
- [ ] Backward compatibility maintained - existing ABI inputs still work
- [ ] No changes to public API signatures that would break existing integrations

---

## Implementation Notes

### Key Design Decisions

1. **Extraction at decode time:** The ABI extraction happens in `CmdDecodeABI` rather than at input validation. This keeps the `RawABI` field containing the original user input for debugging purposes.

2. **Storing extracted ABI:** The `ABI.raw` field stores only the extracted ABI array, not the full wrapped object. This ensures consistent behavior in downstream code that uses the raw ABI string.

3. **Early validation in `abiToJson`:** We validate that wrapped objects have an `abi` field early, but we store the full object. This provides early feedback while allowing the full extraction logic to run later.

4. **Error message clarity:** Error messages explicitly mention "Hardhat/Foundry" to help users understand what formats are supported.

### Edge Cases Handled

- Empty input
- Whitespace-only input
- Valid JSON but wrong type (string, number, boolean)
- Object with `abi` field set to `null`
- Object with `abi` field that is not an array
- Deeply nested `abi` field (not supported - must be top-level)

### GitHub Reference Links

- Main ABI parsing: https://github.com/streamingfast/substreams-codegen/blob/develop/evm-events-calls/abi.go#L26-L38
- Conversation flow: https://github.com/streamingfast/substreams-codegen/blob/develop/evm-events-calls/convo.go#L907-L915
- Test patterns: https://github.com/streamingfast/substreams-codegen/blob/develop/evm-events-calls/abi_test.go
