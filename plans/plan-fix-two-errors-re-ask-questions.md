# Implementation Plan for Codegen Agent

## Objective
Fix a bug in the EVM events-calls codegen flow where providing an incorrect ABI file path twice causes the system to incorrectly re-ask for the contract address instead of continuing to prompt for the correct ABI file path.

## How to Use This Plan
- Follow steps sequentially, completing each before moving to the next
- Verify your work at each checkpoint using the specified commands
- If you encounter ambiguity, refer to the linked source files for context
- Run all verification commands from the project root directory
- Ensure all tests pass before considering a step complete
- Create a pull request when implementation is complete

---

## Problem Analysis

### Observed Behavior
When using the EVM events-calls codegen:
1. User enters a contract address
2. ABI cannot be fetched from block explorer, user chooses "JSON in a local file"
3. User enters an invalid file path (e.g., "test.json") - error correctly shown
4. User enters another invalid file path (e.g., "bob.json")
5. **BUG**: System re-asks "We're tackling the 1st contract. Please enter the contract address" instead of continuing to ask for the ABI file path

### Root Cause Analysis

The bug is caused by the `abiType` field in `BaseContract` being unexported (lowercase), which means it is **not serialized/deserialized** when the conversation state is persisted to JSON.

**File**: `./evm-events-calls/state.go`

Current code (lines 237-247):
```go
type BaseContract struct {
    Name        string          `json:"name,omitempty"`
    TrackEvents bool            `json:"trackEvents"`
    TrackCalls  bool            `json:"trackCalls"`
    RawABI      json.RawMessage `json:"rawAbi,omitempty"`

    abiFetchedInThisSession bool   // Unexported - NOT serialized
    abi                     *ABI   // Unexported - NOT serialized (intentional - complex type)
    emptyABI                bool   // Unexported - NOT serialized
    abiType                 string // Unexported - NOT serialized (BUG!)
}
```

When the client reconnects or the state is rehydrated:
1. The JSON state is deserialized
2. The `abiType` field (which was "file") becomes empty string ""
3. In `NextStep()`, the check `if contract.abiType != ""` fails
4. Instead of going to `AskContractABIType{}`, it goes to `FetchContractABI{}`
5. The fetch fails, returning to `AskContractABIType{}` which appears as re-asking from the start

**Flow in `NextStep()` (lines 99-104)**:
```go
if contract.RawABI == nil {
    // If user already chose how to provide ABI (string/file), skip fetching and ask directly
    if contract.abiType != "" {
        return notifyContext(cmd(AskContractABIType{}))  // Would go here if abiType preserved
    }
    return notifyContext(cmd(FetchContractABI{}))  // Goes here because abiType is ""
}
```

### Why the Bug Manifests After Two Errors

The state is serialized with every `SystemOutput` message sent to the client. If the client reconnects or uses state restoration between error displays, the `abiType` field is lost. The specific "two errors" behavior may be due to client-side reconnection logic or state restoration timing.

---

## Implementation Steps

### Step 1: Export the `abiType` field and add JSON tag

**File**: `./evm-events-calls/state.go`

**Current Code** (lines 237-247):
```go
type BaseContract struct {
    Name        string          `json:"name,omitempty"`
    TrackEvents bool            `json:"trackEvents"`
    TrackCalls  bool            `json:"trackCalls"`
    RawABI      json.RawMessage `json:"rawAbi,omitempty"`

    abiFetchedInThisSession bool
    abi                     *ABI
    emptyABI                bool
    abiType                 string
}
```

**Updated Code**:
```go
type BaseContract struct {
    Name        string          `json:"name,omitempty"`
    TrackEvents bool            `json:"trackEvents"`
    TrackCalls  bool            `json:"trackCalls"`
    RawABI      json.RawMessage `json:"rawAbi,omitempty"`
    AbiType     string          `json:"abiType,omitempty"`

    abiFetchedInThisSession bool
    abi                     *ABI
    emptyABI                bool
}
```

**Rationale**: By exporting `AbiType` (capital A) and adding the JSON tag, the field will be properly serialized/deserialized when the state is persisted and restored.

### Step 2: Update all references from `abiType` to `AbiType`

**File**: `./evm-events-calls/state.go`

No additional references in this file - the field is only used in `convo.go`.

**File**: `./evm-events-calls/convo.go`

Update all references to use the exported field name:

**Location 1** (line 101):
```go
// Before
if contract.abiType != "" {

// After
if contract.AbiType != "" {
```

**Location 2** (line 281):
```go
// Before
if contract.abiType == "string" {

// After
if contract.AbiType == "string" {
```

**Location 3** (line 283):
```go
// Before
} else if contract.abiType == "file" {

// After
} else if contract.AbiType == "file" {
```

**Location 4** (line 300):
```go
// Before
contract.abiType = msg.Value

// After
contract.AbiType = msg.Value
```

**Location 5** (line 303):
```go
// Before
if contract.abiType == "string" {

// After
if contract.AbiType == "string" {
```

### Step 3: Add a unit test to verify retry behavior persists across state restoration

**File**: `./evm-events-calls/convo_test.go`

Add the following test at the end of the file:

```go
func TestABIFileErrorRetryPersistsAcrossStateRestore(t *testing.T) {
    // This test verifies that after choosing "file" as ABI type,
    // the choice persists even if state is serialized/deserialized
    // (simulating client reconnection)

    conv := New()
    conv.SetFactory(&codegen.MsgWrapFactory{})
    p := conv.(*Convo).State

    // Setup: project name, chain, contract with address
    p.Name = "test_project"
    p.ChainName = "mainnet"
    p.Contracts = append(p.Contracts, &Contract{
        BaseContract: BaseContract{},
        Address:      "0x1f98431c8ad98523631ae4a59f267346ea31f984",
    })
    p.currentContractIdx = 0

    // User chooses "file" as ABI type
    next := conv.Update(InputContractABIType{UserInput_Selection: pbconvo.UserInput_Selection{Value: "file"}})
    assert.Equal(t, AskContractABIFile{}, next())
    assert.Equal(t, "file", p.Contracts[0].AbiType)

    // Simulate state serialization and deserialization (client reconnection)
    stateJSON, err := json.Marshal(p)
    require.NoError(t, err)

    // Create new conversation and restore state
    conv2 := New()
    conv2.SetFactory(&codegen.MsgWrapFactory{})
    p2 := conv2.(*Convo).State
    err = json.Unmarshal(stateJSON, p2)
    require.NoError(t, err)

    // Verify AbiType was preserved
    require.Len(t, p2.Contracts, 1)
    assert.Equal(t, "file", p2.Contracts[0].AbiType, "AbiType should be preserved after state restoration")

    // Simulate first file error
    errorMsg := "could not read file \"test.json\": no such file or directory"
    next = conv2.Update(InputContractABIFile{UserInput_LocalFile: pbconvo.UserInput_LocalFile{
        Error: &errorMsg,
    }})

    // Should show error and re-ask for file, NOT go back to contract address
    seq := next().(loop.SeqMsg)
    require.Len(t, seq, 2)

    // First is error message
    msg := seq[0]().(*pbconvo.SystemOutput)
    assert.Contains(t, msg.GetMessage().Markdown, "File not found")

    // Second should be AskContractABIFile, NOT AskContractAddress
    assert.Equal(t, AskContractABIFile{}, seq[1](), "Should ask for file again, not contract address")

    // Simulate second file error
    errorMsg2 := "could not read file \"bob.json\": no such file or directory"
    next = conv2.Update(InputContractABIFile{UserInput_LocalFile: pbconvo.UserInput_LocalFile{
        Error: &errorMsg2,
    }})

    // Should STILL show error and re-ask for file
    seq = next().(loop.SeqMsg)
    require.Len(t, seq, 2)

    // First is error message
    msg = seq[0]().(*pbconvo.SystemOutput)
    assert.Contains(t, msg.GetMessage().Markdown, "File not found")

    // Second should be AskContractABIFile, NOT AskContractAddress
    assert.Equal(t, AskContractABIFile{}, seq[1](), "Should ask for file again on second error, not contract address")
}
```

Add the required import at the top of the file if not present:
```go
import (
    "encoding/json"
    // ... other imports
)
```

---

## Summary of Changes

### Files Modified

1. **`./evm-events-calls/state.go`**:
   - Export `AbiType` field (rename from `abiType` to `AbiType`)
   - Add JSON tag `json:"abiType,omitempty"`

2. **`./evm-events-calls/convo.go`**:
   - Update 5 references from `abiType` to `AbiType`:
     - Line 101: `contract.abiType` -> `contract.AbiType`
     - Line 281: `contract.abiType` -> `contract.AbiType`
     - Line 283: `contract.abiType` -> `contract.AbiType`
     - Line 300: `contract.abiType` -> `contract.AbiType`
     - Line 303: `contract.abiType` -> `contract.AbiType`

3. **`./evm-events-calls/convo_test.go`**:
   - Add new test `TestABIFileErrorRetryPersistsAcrossStateRestore`
   - Add `encoding/json` import if not present

### Behavior After Fix

| Scenario | Before Fix | After Fix |
|----------|-----------|-----------|
| First file error | Shows error, re-asks for file | Same |
| Second file error (same session) | May go back to contract address | Shows error, re-asks for file |
| File error after state restore | Goes back to contract address | Shows error, re-asks for file |

---

## Verification Steps

After implementing each component:

1. **Run the new test specifically**
   ```bash
   go test ./evm-events-calls/... -v -run TestABIFileErrorRetryPersistsAcrossStateRestore
   ```
   Expected: Test passes, demonstrating that AbiType persists across state restoration

2. **Run all EVM events-calls tests**
   ```bash
   go test ./evm-events-calls/... -v
   ```
   Expected: All tests pass, including existing tests

3. **Run the full test suite**
   ```bash
   go test ./... -v
   ```
   Expected: All tests pass, no regressions

4. **Build verification**
   ```bash
   go build ./cmd/...
   ```
   Expected: Clean build with no warnings or errors

5. **Manual testing** (optional but recommended)
   - Run the codegen server locally
   - Start an EVM events-calls conversation
   - Enter a contract address
   - When ABI fetch fails, choose "JSON in a local file"
   - Enter an invalid file path - verify error is shown
   - Enter another invalid file path - verify it asks for file again (not contract address)

## Project Standards Checklist
- [ ] Code follows Go naming conventions (exported field uses PascalCase)
- [ ] JSON tag follows existing patterns in the codebase (`json:"camelCase,omitempty"`)
- [ ] Error messages are descriptive and actionable
- [ ] Test uses table-driven approach where applicable
- [ ] Test uses `require.NoError(t, err)` for error checking
- [ ] Test helper functions call `t.Helper()` if any are added

---

## Key Implementation Details

### State Serialization Pattern

The conversation state is serialized to JSON and included in every `SystemOutput` message. This allows clients to restore the conversation state on reconnection. See:
- https://github.com/streamingfast/substreams-codegen/blob/develop/msgwrap.go#L74-L84

### Why Other Unexported Fields Are OK

- `abiFetchedInThisSession`: Ephemeral flag for current session only, not needed after reconnect
- `abi`: Complex pointer type that gets rebuilt from `RawABI` via `RunDecodeContractABI`
- `emptyABI`: Transient flag that triggers re-asking for contract address (intentional behavior)

The `abiType` field is different because it represents a user choice that should persist across reconnections.

### Reference: Similar Fix in Starknet Events

The same pattern exists in the starknet-events module. If this fix is successful, consider auditing other codegen modules for similar issues with unexported fields that represent user choices.

### Related Error Handling

The error display logic in `InputContractABIFile` correctly returns to `AskContractABIFile{}`:

https://github.com/streamingfast/substreams-codegen/blob/develop/evm-events-calls/convo.go#L349-L352

```go
case InputContractABIFile:
    if msg.Error != nil && *msg.Error != "" {
        friendlyErr := codegen.MapClientSideErrorToMessage(fmt.Errorf("%s", *msg.Error))
        return loop.Seq(c.Msg().Messagef("Unable to read ABI: %s", friendlyErr).Cmd(), cmd(AskContractABIFile{}))
    }
```

This logic is correct and does not need to be changed. The fix is solely about ensuring `AbiType` persists across state restoration.
