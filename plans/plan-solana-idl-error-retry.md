# Implementation Plan: Fix Solana Anchor IDL Error Retry

## Objective

Fix the `sol-anchor` generator to allow users to retry when providing an invalid IDL file or JSON string, instead of terminating the conversation abruptly. This makes the behavior consistent with EVM generators.

## How to Use This Plan

- Follow steps sequentially, completing each before moving to the next
- Verify your work at each checkpoint using the specified commands
- Run all verification commands from the project root directory
- Ensure all tests pass before considering a step complete

---

## Problem Statement

### Current Behavior (Broken)

When a user provides an invalid IDL (via file or JSON string) to the `sol-anchor` generator:
1. ✅ File read errors → Shows error, allows retry (CORRECT)
2. ❌ Invalid JSON → **Quits conversation immediately** (INCORRECT)
3. ❌ Empty content → Passed to parser, which fails and quits (INCORRECT)

### Expected Behavior

Match the EVM generators' pattern:
1. ✅ File read errors → Show error, ask again
2. ✅ Invalid JSON → Show error, ask again
3. ✅ Empty content → Show error, ask again

### Root Cause

**File**: `./sol-anchor/convo.go` lines 205-209

The `inputIDLStep()` function calls `loop.Quit()` when JSON parsing fails:

```go
func inputIDLStep(c *Convo, msgValue string) loop.Cmd {
    idl, err := createIDLFromJSON(msgValue)
    if err != nil {
        return loop.Quit(fmt.Errorf("could not decode IDL"))  // ❌ WRONG
    }
    // ...
}
```

**Should be**: Show error message and return to the appropriate Ask step (either `AskIDLFile{}` or `AskIDLJSON{}` depending on input method).

---

## Implementation Steps

### Step 1: Add helper method to determine retry step

**File**: `./sol-anchor/convo.go`

Add a helper method after `inputIDLStep()` to determine which Ask command to use:

```go
// askIDLAgain returns the appropriate Ask command based on the current IDL input format
func (c *Convo) askIDLAgain() loop.Cmd {
    if c.State.IdlFormat == "file" {
        return cmd(AskIDLFile{})
    }
    return cmd(AskIDLJSON{})
}
```

**Rationale**: We need to know whether to ask for a file path again or JSON string again, depending on how the user chose to provide the IDL initially.

### Step 2: Update `inputIDLStep` to handle errors gracefully

**File**: `./sol-anchor/convo.go`

**Current code (lines 205-221)**:
```go
func inputIDLStep(c *Convo, msgValue string) loop.Cmd {
    idl, err := createIDLFromJSON(msgValue)
    if err != nil {
        return loop.Quit(fmt.Errorf("could not decode IDL"))
    }
    if idl.Metadata.Name == "" {
        idl.Metadata.Name = c.State.Name // we need a name so anchor can compile
    }

    c.State.idl = idl
    c.State.IdlString = msgValue

    descString := descriptionFromIDL(idl)

    peekIDL := c.Msg().Message(descString).Cmd()
    return loop.Seq(peekIDL, cmd(AskConfirmIDL{}))
}
```

**New code**:
```go
func inputIDLStep(c *Convo, msgValue string) loop.Cmd {
    // Check for empty input first
    trimmed := strings.TrimSpace(msgValue)
    if trimmed == "" {
        return loop.Seq(
            c.Msg().Messagef("The IDL content is empty. Please provide a valid Anchor JSON IDL.").Cmd(),
            c.askIDLAgain(),
        )
    }

    idl, err := createIDLFromJSON(msgValue)
    if err != nil {
        // Show a helpful error message and allow retry
        errorMsg := fmt.Sprintf("Could not decode IDL: %v\n\nPlease ensure your IDL is valid JSON in Anchor format.", err)
        return loop.Seq(
            c.Msg().Messagef(errorMsg).Cmd(),
            c.askIDLAgain(),
        )
    }

    if idl.Metadata.Name == "" {
        idl.Metadata.Name = c.State.Name // we need a name so anchor can compile
    }

    c.State.idl = idl
    c.State.IdlString = msgValue

    descString := descriptionFromIDL(idl)

    peekIDL := c.Msg().Message(descString).Cmd()
    return loop.Seq(peekIDL, cmd(AskConfirmIDL{}))
}
```

**Add import** at the top of the file:
```go
import (
    "strings"  // ADD THIS if not already present
    // ... other imports
)
```

### Step 3: Remove debug print statements

**File**: `./sol-anchor/convo.go`

**Location**: Lines 223-229 in `createIDLFromJSON()`

**Current code**:
```go
func createIDLFromJSON(text string) (*IDL, error) {
    idl := &IDL{}
    err := json.Unmarshal([]byte(text), &idl)
    if err != nil {
        fmt.Println("Error unmarshaling JSON:", err)  // ❌ Remove this
        return nil, err
    }
    idl.MoveEventsIfNecessary()

    return idl, nil
}
```

**New code**:
```go
func createIDLFromJSON(text string) (*IDL, error) {
    idl := &IDL{}
    err := json.Unmarshal([]byte(text), &idl)
    if err != nil {
        return nil, fmt.Errorf("failed to parse IDL JSON: %w", err)
    }
    idl.MoveEventsIfNecessary()

    return idl, nil
}
```

**Rationale**: Remove debug prints and wrap errors properly for better error messages.

### Step 4: Add tests for error retry scenarios

**File**: `./sol-anchor/convo_test.go`

Add test cases to verify retry behavior:

```go
func TestIDLErrorRetryBehavior(t *testing.T) {
    t.Helper()

    tests := []struct {
        name           string
        idlFormat      string
        idlInput       string
        expectErrorMsg string
        expectAskAgain interface{}
    }{
        {
            name:           "empty IDL from file",
            idlFormat:      "file",
            idlInput:       "   ",
            expectErrorMsg: "empty",
            expectAskAgain: AskIDLFile{},
        },
        {
            name:           "empty IDL from JSON string",
            idlFormat:      "string",
            idlInput:       "",
            expectErrorMsg: "empty",
            expectAskAgain: AskIDLJSON{},
        },
        {
            name:           "invalid JSON from file",
            idlFormat:      "file",
            idlInput:       `{invalid json}`,
            expectErrorMsg: "Could not decode IDL",
            expectAskAgain: AskIDLFile{},
        },
        {
            name:           "invalid JSON from string",
            idlFormat:      "string",
            idlInput:       `{"abi": not valid}`,
            expectErrorMsg: "Could not decode IDL",
            expectAskAgain: AskIDLJSON{},
        },
        {
            name:           "valid JSON but not IDL format from file",
            idlFormat:      "file",
            idlInput:       `{"some": "random", "json": true}`,
            expectErrorMsg: "Could not decode IDL",
            expectAskAgain: AskIDLFile{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Helper()

            conv := New()
            conv.SetFactory(&codegen.MsgWrapFactory{})
            p := conv.(*Convo).State

            // Setup: we're at the point of IDL input
            p.Name = "test-project"
            p.ChainName = "solana-mainnet"
            p.IdlFormat = tt.idlFormat

            // Simulate file or string input with invalid content
            var next loop.Cmd
            if tt.idlFormat == "file" {
                next = conv.Update(InputIDLFile{
                    UserInput_LocalFile: pbconvo.UserInput_LocalFile{
                        Value: []byte(tt.idlInput),
                    },
                })
            } else {
                next = conv.Update(InputIDLJSON{
                    UserInput_TextInput: pbconvo.UserInput_TextInput{
                        Value: tt.idlInput,
                    },
                })
            }

            // Should get a sequence: error message + ask again
            seq := next().(loop.SeqMsg)
            require.Len(t, seq, 2, "should return sequence of 2 commands")

            // First command should be error message
            msg := seq[0]().(*pbconvo.SystemOutput)
            assert.Contains(t, msg.GetMessage().Markdown, tt.expectErrorMsg, "error message should contain expected text")

            // Second command should be the Ask command
            askCmd := seq[1]()
            assert.IsType(t, tt.expectAskAgain, askCmd, "should ask for IDL again")
        })
    }
}

func TestIDLSuccessfulInput(t *testing.T) {
    t.Helper()

    conv := New()
    conv.SetFactory(&codegen.MsgWrapFactory{})
    p := conv.(*Convo).State

    // Setup
    p.Name = "test-project"
    p.ChainName = "solana-mainnet"
    p.IdlFormat = "string"

    // Provide valid minimal IDL
    validIDL := `{
        "version": "0.1.0",
        "name": "test",
        "metadata": {"address": "GovaE4iu227srtG2s3tZzB4RmWBzw8sTwrCLZz7kN7rY"},
        "instructions": [],
        "accounts": [],
        "types": [],
        "events": []
    }`

    next := conv.Update(InputIDLJSON{
        UserInput_TextInput: pbconvo.UserInput_TextInput{
            Value: validIDL,
        },
    })

    // Should get a sequence: preview message + AskConfirmIDL
    seq := next().(loop.SeqMsg)
    require.Len(t, seq, 2, "should return sequence of 2 commands")

    // Second command should be AskConfirmIDL
    askCmd := seq[1]()
    assert.IsType(t, AskConfirmIDL{}, askCmd, "should ask for IDL confirmation")
}
```

---

## File Modification Summary

| File | Changes |
|------|---------|
| `./sol-anchor/convo.go` | Add `askIDLAgain()` helper, update `inputIDLStep()` error handling, improve `createIDLFromJSON()` error wrapping |
| `./sol-anchor/convo_test.go` | Add `TestIDLErrorRetryBehavior` and `TestIDLSuccessfulInput` |

---

## Verification Steps

### 1. Unit Tests

```bash
# Run sol-anchor tests
go test ./sol-anchor/... -v

# Run specifically the new tests
go test ./sol-anchor/... -v -run TestIDLError
go test ./sol-anchor/... -v -run TestIDLSuccessful
```

**Expected**: All tests pass, including new retry behavior tests.

### 2. Build Verification

```bash
# Build all commands
go build ./cmd/...

# Verify no compilation errors
echo $?  # Should be 0
```

**Expected**: Clean build with no warnings or errors.

### 3. Backward Compatibility Check

```bash
# Run the full test suite
go test ./... -v
```

**Expected**: All existing tests continue to pass.

---

## Success Criteria

- [ ] `askIDLAgain()` helper method added
- [ ] `inputIDLStep()` handles empty input gracefully
- [ ] `inputIDLStep()` shows error and allows retry on JSON parse failures
- [ ] `createIDLFromJSON()` has proper error wrapping (no debug prints)
- [ ] New tests verify retry behavior for all error cases
- [ ] All existing tests still pass
- [ ] Build succeeds with no errors

---

## Testing Scenarios

### Error Cases (Should Allow Retry)

1. **Empty file content**
   - Input: File with only whitespace
   - Expected: Error message → `AskIDLFile{}`

2. **Empty JSON string**
   - Input: Empty string or whitespace
   - Expected: Error message → `AskIDLJSON{}`

3. **Invalid JSON syntax**
   - Input: `{invalid json`
   - Expected: Error message → Ask again (file or string depending on format)

4. **Valid JSON, wrong structure**
   - Input: `{"random": "object"}`
   - Expected: Error message → Ask again

### Success Case

5. **Valid IDL**
   - Input: Proper Anchor IDL JSON
   - Expected: Show IDL preview → `AskConfirmIDL{}`

---

## Implementation Notes

### Key Design Decisions

1. **Helper method `askIDLAgain()`**: Centralizes the logic for determining which Ask command to use, making the code more maintainable.

2. **Empty input check first**: Check for empty/whitespace input before attempting JSON parsing to provide a more specific error message.

3. **Error message clarity**: Include the underlying error in the message so users understand what went wrong with their IDL.

4. **Consistent with EVM pattern**: Follows the same error handling pattern used in `evm-events-calls/convo.go` for ABI input.

### Edge Cases Handled

- Empty input (whitespace only)
- Invalid JSON syntax
- Valid JSON but wrong structure
- Missing required IDL fields
- File read errors (already handled correctly)

---

## Project Standards Checklist

- [ ] Code follows Go conventions and existing patterns in the repository
- [ ] Error messages are descriptive and actionable
- [ ] Helper method has appropriate documentation comment
- [ ] Table-driven tests using testify assert/require
- [ ] Uses `require.NoError(t, err)` for all error checking in tests
- [ ] Test helpers call `t.Helper()` as first statement
- [ ] No debug print statements (`fmt.Println`) in production code
- [ ] Error wrapping uses `fmt.Errorf` with `%w` verb
