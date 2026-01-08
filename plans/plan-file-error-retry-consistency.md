# Implementation Plan for Codegen Agent

## Objective
Fix the inconsistent error handling when users provide erroneous files during conversation flows in `substreams-codegen`. Make all generators (EVM and Solana) behave consistently by allowing retries when file inputs fail, rather than terminating the conversation abruptly.

## How to Use This Plan
- Follow steps sequentially, completing each before moving to the next
- Verify your work at each checkpoint using the specified commands
- If you encounter ambiguity, refer to the linked source files for context
- Run all verification commands from the project root directory
- Ensure all tests pass before considering a step complete
- Create a pull request when implementation is complete

---

## Problem Analysis

### Current Behavior

**EVM Generators (`evm-events-calls`)**:
When a user provides an erroneous ABI file (via `InputContractABIFile`):
1. If the file has a read error, it shows an error message and asks for the file again (`AskContractABIFile{}`):
   ```go
   // ./evm-events-calls/convo.go lines 340-342
   case InputContractABIFile:
       if msg.Error != nil && *msg.Error != "" {
           return loop.Seq(c.Msg().Messagef("The ABI file couldn't be read correctly: %q", *msg.Error).Cmd(), cmd(AskContractABIFile{}))
       }
   ```
2. If the content is empty, it sets `emptyABI = true` and goes back in the flow to ask for the contract address again
3. If the JSON is invalid, it shows an error message and asks for the file again
4. **Allows retries** - the conversation continues

**Solana Generators (`sol-anchor`)**:
When a user provides an erroneous IDL file (via `InputIDLFile`):
1. If the file has a read error, it shows an error message and asks for the file again:
   ```go
   // ./sol-anchor/convo.go lines 140-142
   case InputIDLFile:
       if msg.Error != nil && *msg.Error != "" {
           return loop.Seq(c.Msg().Messagef("The IDL file couldn't be read correctly: %q", *msg.Error).Cmd(), cmd(AskIDLFile{}))
       }
   ```
2. If JSON parsing fails, it **immediately quits** with an error:
   ```go
   // ./sol-anchor/convo.go lines 205-209
   func inputIDLStep(c *Convo, msgValue string) loop.Cmd {
       idl, err := createIDLFromJSON(msgValue)
       if err != nil {
           return loop.Quit(fmt.Errorf("could not decode IDL"))
       }
       // ...
   }
   ```

### Root Cause

The inconsistency is in `inputIDLStep()` function in `./sol-anchor/convo.go`. When JSON parsing fails, it calls `loop.Quit()` which terminates the entire conversation, instead of:
1. Displaying an error message to the user
2. Returning to the appropriate Ask step to allow the user to try again

### Affected Generators

| Generator | File Input Type | Retry on Error | Current Behavior |
|-----------|----------------|----------------|------------------|
| `evm-events-calls` | `InputContractABIFile` | Yes | Shows error, asks again |
| `evm-events-calls` | `InputContractABIString` | Yes | Shows error, asks again |
| `starknet-events` | `InputContractABI` | Yes | Shows error, asks again |
| `sol-anchor` | `InputIDLFile` | Partial | Only for read errors, quits on parse errors |
| `sol-anchor` | `InputIDLJSON` | No | Quits on parse errors |

---

## Implementation Steps

### Step 1: Fix `sol-anchor` JSON parsing error handling

**File**: `./sol-anchor/convo.go`

**Current Code (lines 205-221)**:
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

**Updated Code**:
```go
func inputIDLStep(c *Convo, msgValue string) loop.Cmd {
    // Handle empty input - go back to ask for input
    if strings.TrimSpace(msgValue) == "" {
        return loop.Seq(
            c.Msg().Message("The IDL content is empty. Please provide a valid Anchor IDL.").Cmd(),
            c.askIDLAgain(),
        )
    }

    idl, err := createIDLFromJSON(msgValue)
    if err != nil {
        // Show error message and ask for IDL again instead of quitting
        return loop.Seq(
            c.Msg().Messagef("The IDL content is not valid JSON: %s. Please try again.", err).Cmd(),
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

// askIDLAgain returns the appropriate Ask command based on the current IdlFormat
func (c *Convo) askIDLAgain() loop.Cmd {
    if c.State.IdlFormat == "file" {
        return cmd(AskIDLFile{})
    }
    return cmd(AskIDLJSON{})
}
```

**Note**: Add `"strings"` to imports if not already present.

### Step 2: Handle empty file content in `sol-anchor`

**File**: `./sol-anchor/convo.go`

The `InputIDLFile` case already handles read errors properly. However, we should also handle empty file content like the EVM generators do.

**Current Code (lines 140-145)**:
```go
case InputIDLFile:
    if msg.Error != nil && *msg.Error != "" {
        return loop.Seq(c.Msg().Messagef("The IDL file couldn't be read correctly: %q", *msg.Error).Cmd(), cmd(AskIDLFile{}))
    }

    return inputIDLStep(c, string(msg.Value))
```

This is already correct since `inputIDLStep` will now handle empty content. No changes needed here.

### Step 3: Update `createIDLFromJSON` to provide better error messages

**File**: `./sol-anchor/convo.go`

**Current Code (lines 223-233)**:
```go
func createIDLFromJSON(text string) (*IDL, error) {
    idl := &IDL{}
    err := json.Unmarshal([]byte(text), &idl)
    if err != nil {
        fmt.Println("Error unmarshaling JSON:", err)
        return nil, err
    }
    idl.MoveEventsIfNecessary()

    return idl, nil
}
```

**Updated Code**:
```go
func createIDLFromJSON(text string) (*IDL, error) {
    idl := &IDL{}
    err := json.Unmarshal([]byte(text), &idl)
    if err != nil {
        return nil, fmt.Errorf("JSON parsing error: %w", err)
    }
    idl.MoveEventsIfNecessary()

    return idl, nil
}
```

Remove the `fmt.Println` debugging statement since errors are now properly returned to the user through the conversation.

### Step 4: Add tests for error retry behavior

**File**: `./sol-anchor/convo_test.go`

Add tests to verify the retry behavior works correctly.

```go
func TestIDLParseErrorRetry(t *testing.T) {
    conv := New()
    conv.SetFactory(&codegen.MsgWrapFactory{})
    p := conv.(*Convo).State

    // Setup state
    p.Name = "test_project"
    p.ChainName = "solana-mainnet-beta"
    p.IdlFormat = "string"

    // Provide invalid JSON
    next := conv.Update(InputIDLJSON{
        UserInput_TextInput: pbconvo.UserInput_TextInput{
            Value: "{ invalid json }",
        },
    })

    // Should return a sequence with error message and ask for IDL again
    seq, ok := next().(loop.SeqMsg)
    require.True(t, ok, "expected SeqMsg")
    require.Len(t, seq, 2, "expected 2 commands in sequence")

    // First should be error message
    msg := seq[0]().(*pbconvo.SystemOutput)
    assert.Contains(t, msg.GetMessage().Markdown, "not valid JSON")

    // Second should ask for IDL again
    assert.Equal(t, AskIDLJSON{}, seq[1]())
}

func TestIDLEmptyInputRetry(t *testing.T) {
    conv := New()
    conv.SetFactory(&codegen.MsgWrapFactory{})
    p := conv.(*Convo).State

    // Setup state
    p.Name = "test_project"
    p.ChainName = "solana-mainnet-beta"
    p.IdlFormat = "string"

    // Provide empty input
    next := conv.Update(InputIDLJSON{
        UserInput_TextInput: pbconvo.UserInput_TextInput{
            Value: "",
        },
    })

    // Should return a sequence with message and ask for IDL again
    seq, ok := next().(loop.SeqMsg)
    require.True(t, ok, "expected SeqMsg")
    require.Len(t, seq, 2, "expected 2 commands in sequence")

    // First should be message about empty content
    msg := seq[0]().(*pbconvo.SystemOutput)
    assert.Contains(t, msg.GetMessage().Markdown, "empty")

    // Second should ask for IDL again
    assert.Equal(t, AskIDLJSON{}, seq[1]())
}

func TestIDLFileReadErrorRetry(t *testing.T) {
    conv := New()
    conv.SetFactory(&codegen.MsgWrapFactory{})
    p := conv.(*Convo).State

    // Setup state
    p.Name = "test_project"
    p.ChainName = "solana-mainnet-beta"
    p.IdlFormat = "file"

    // Provide file with read error
    errorMsg := "file not found"
    next := conv.Update(InputIDLFile{
        UserInput_LocalFile: pbconvo.UserInput_LocalFile{
            Error: &errorMsg,
        },
    })

    // Should return a sequence with error message and ask for IDL file again
    seq, ok := next().(loop.SeqMsg)
    require.True(t, ok, "expected SeqMsg")
    require.Len(t, seq, 2, "expected 2 commands in sequence")

    // First should be error message
    msg := seq[0]().(*pbconvo.SystemOutput)
    assert.Contains(t, msg.GetMessage().Markdown, "file not found")

    // Second should ask for IDL file again
    assert.Equal(t, AskIDLFile{}, seq[1]())
}

func TestIDLFileInvalidJSONRetry(t *testing.T) {
    conv := New()
    conv.SetFactory(&codegen.MsgWrapFactory{})
    p := conv.(*Convo).State

    // Setup state
    p.Name = "test_project"
    p.ChainName = "solana-mainnet-beta"
    p.IdlFormat = "file"

    // Provide file with invalid JSON content
    next := conv.Update(InputIDLFile{
        UserInput_LocalFile: pbconvo.UserInput_LocalFile{
            Value: []byte("not json at all"),
        },
    })

    // Should return a sequence with error message and ask for IDL file again
    seq, ok := next().(loop.SeqMsg)
    require.True(t, ok, "expected SeqMsg")
    require.Len(t, seq, 2, "expected 2 commands in sequence")

    // First should be error message about JSON
    msg := seq[0]().(*pbconvo.SystemOutput)
    assert.Contains(t, msg.GetMessage().Markdown, "not valid JSON")

    // Second should ask for IDL file again
    assert.Equal(t, AskIDLFile{}, seq[1]())
}
```

---

## Summary of Changes

### Files Modified

1. **`./sol-anchor/convo.go`**:
   - Update `inputIDLStep()` to show error message and retry instead of quitting
   - Add `askIDLAgain()` helper method to return appropriate Ask command
   - Update `createIDLFromJSON()` to remove debug print and wrap error properly
   - Add `strings` import if needed

2. **`./sol-anchor/convo_test.go`**:
   - Add `TestIDLParseErrorRetry` test
   - Add `TestIDLEmptyInputRetry` test
   - Add `TestIDLFileReadErrorRetry` test
   - Add `TestIDLFileInvalidJSONRetry` test

### Behavior After Fix

| Generator | File Input Type | Retry on Error | New Behavior |
|-----------|----------------|----------------|--------------|
| `evm-events-calls` | `InputContractABIFile` | Yes | No change |
| `evm-events-calls` | `InputContractABIString` | Yes | No change |
| `starknet-events` | `InputContractABI` | Yes | No change |
| `sol-anchor` | `InputIDLFile` | **Yes** | Shows error, asks again |
| `sol-anchor` | `InputIDLJSON` | **Yes** | Shows error, asks again |

---

## Verification Steps

After implementing each component:

1. **Unit Tests**
   ```bash
   go test ./sol-anchor/... -v -run TestIDL
   ```
   Expected: All IDL-related tests pass, including new retry tests

2. **Full Test Suite**
   ```bash
   go test ./... -v
   ```
   Expected: All tests pass, no regressions

3. **Build Verification**
   ```bash
   go build ./cmd/...
   ```
   Expected: Clean build with no warnings

4. **Manual Testing** (optional)
   - Run the codegen server locally
   - Use the CLI to start a sol-anchor conversation
   - Provide an invalid JSON file and verify you can retry
   - Provide an empty file and verify you can retry
   - Provide a valid IDL file and verify the flow continues normally

## Project Standards Checklist
- [ ] Code follows Go conventions and existing patterns in the repository
- [ ] Error messages are descriptive and actionable (tell user what went wrong and that they can try again)
- [ ] Logging is consistent with the project's logging framework (removed debug `fmt.Println`)
- [ ] New functions/methods have appropriate documentation
- [ ] Table-driven tests follow project conventions with `require.NoError` and `assert`
- [ ] Uses `t.Helper()` in test helpers as per CLAUDE.md

---

## Key Implementation Details

### Error Handling Pattern

The EVM generators use this consistent pattern for file input errors:

```go
case InputContractABIFile:
    // 1. Check for read errors first
    if msg.Error != nil && *msg.Error != "" {
        return loop.Seq(
            c.Msg().Messagef("The ABI file couldn't be read correctly: %q", *msg.Error).Cmd(),
            cmd(AskContractABIFile{}),  // Ask again
        )
    }

    // 2. Check for empty content
    if string(msg.Value) == "" {
        contract.emptyABI = true
        return c.NextStep()  // Go back in flow
    }

    // 3. Validate content
    rawAbi, err := abiToJson(string(msg.Value))
    if err != nil {
        return loop.Seq(
            c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(),
            cmd(AskContractABIFile{}),  // Ask again
        )
    }

    // 4. Success path
    contract.RawABI = rawAbi
    return c.NextStep()
```

Reference: https://github.com/streamingfast/substreams-codegen/blob/develop/evm-events-calls/convo.go#L340-L362

The Solana generator should follow the same pattern.

### Loop Package Commands

The `loop` package provides these key commands:
- `loop.Seq(cmds...)` - Execute commands in sequence
- `loop.Quit(err)` - Terminate the conversation with an error
- `cmd(AskXXX{})` - Trigger an Ask message to prompt user input

Reference: https://github.com/streamingfast/substreams-codegen/blob/develop/loop/cmd.go

### Message Wrapping

Use `c.Msg().Messagef()` to format user-facing messages. This properly wraps the message for the conversation protocol.

Reference: https://github.com/streamingfast/substreams-codegen/blob/develop/msgwrap.go
