# Implementation Plan for Codegen Agent

## Objective
Restore the lost ABI preview feature in the EVM events-calls generator. When a user provides an ABI (either via manual input or file), the system should display a color-highlighted rendering of the parsed ABI showing events, calls, and their fields before proceeding.

## How to Use This Plan
- Follow steps sequentially, completing each before moving to the next
- Verify your work at each checkpoint using the specified commands
- If you encounter ambiguity, refer to the linked source files for context
- Run all verification commands from the project root directory
- Ensure all tests pass before considering a step complete
- Create a pull request when implementation is complete

---

## Root Cause Analysis

### The Problem
The ABI preview feature is **conditionally controlled** by a flag `abiFetchedInThisSession` on the contract. The preview is ONLY shown when this flag is `true`:

```go
// In evm-events-calls/convo.go, lines 494-522
if !contract.abiFetchedInThisSession {
    return c.NextStep()  // <-- Skips the preview!
}

// the 'printf' usage in the Go template below is a hack because we can't do
// arithmetics in the template, it means '+1'
peekABI := c.Msg().MessageTpl(`Ok, here's what the ABI would produce:
...
```

### When Was This Broken?
The issue was introduced in commit `58e9e94` (March 20, 2025) titled "evm-events-calls: Fix reading from file & new path to avoid usage of ABI (#26)".

This commit introduced separate handlers for string and file ABI input (`InputContractABIString` and `InputContractABIFile`) but **forgot to set `contract.abiFetchedInThisSession = true`** in these new handlers.

### Current Behavior
- **API-fetched ABI**: Sets `abiFetchedInThisSession = true` -> Preview IS shown
- **Manual string input**: Does NOT set the flag -> Preview is SKIPPED
- **File input**: Does NOT set the flag -> Preview is SKIPPED

### Expected Behavior
ALL ABI inputs should show the preview, regardless of source (API, string, or file).

---

## Implementation Steps

### Step 1: Set `abiFetchedInThisSession` Flag for String Input

**File**: `./evm-events-calls/convo.go`

**Location**: Find the `case InputContractABIString:` handler (around line 310-328)

**Current Code**:
```go
case InputContractABIString:
    contract := c.contextContract()
    if contract == nil {
        return QuitInvalidContext
    }

    if msg.Value == "" {
        contract.emptyABI = true
        return c.NextStep()
    }

    rawAbi, err := abiToJson(msg.Value)
    if err != nil {
        return loop.Seq(c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskContractABIString{}))
    }

    contract.RawABI = rawAbi

    return c.NextStep()
```

**Change Required**: Add `contract.abiFetchedInThisSession = true` before returning:

```go
case InputContractABIString:
    contract := c.contextContract()
    if contract == nil {
        return QuitInvalidContext
    }

    if msg.Value == "" {
        contract.emptyABI = true
        return c.NextStep()
    }

    rawAbi, err := abiToJson(msg.Value)
    if err != nil {
        return loop.Seq(c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskContractABIString{}))
    }

    contract.RawABI = rawAbi
    contract.abiFetchedInThisSession = true  // ADD THIS LINE

    return c.NextStep()
```

### Step 2: Set `abiFetchedInThisSession` Flag for File Input

**File**: `./evm-events-calls/convo.go`

**Location**: Find the `case InputContractABIFile:` handler (around line 340-362)

**Current Code**:
```go
case InputContractABIFile:
    if msg.Error != nil && *msg.Error != "" {
        return loop.Seq(c.Msg().Messagef("The ABI file couldn't be read correctly: %q", *msg.Error).Cmd(), cmd(AskContractABIFile{}))
    }

    contract := c.contextContract()
    if contract == nil {
        return QuitInvalidContext
    }

    if string(msg.Value) == "" {
        contract.emptyABI = true
        return c.NextStep()
    }

    rawAbi, err := abiToJson(string(msg.Value))
    if err != nil {
        return loop.Seq(c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskContractABIFile{}))
    }

    contract.RawABI = rawAbi

    return c.NextStep()
```

**Change Required**: Add `contract.abiFetchedInThisSession = true` before returning:

```go
case InputContractABIFile:
    if msg.Error != nil && *msg.Error != "" {
        return loop.Seq(c.Msg().Messagef("The ABI file couldn't be read correctly: %q", *msg.Error).Cmd(), cmd(AskContractABIFile{}))
    }

    contract := c.contextContract()
    if contract == nil {
        return QuitInvalidContext
    }

    if string(msg.Value) == "" {
        contract.emptyABI = true
        return c.NextStep()
    }

    rawAbi, err := abiToJson(string(msg.Value))
    if err != nil {
        return loop.Seq(c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskContractABIFile{}))
    }

    contract.RawABI = rawAbi
    contract.abiFetchedInThisSession = true  // ADD THIS LINE

    return c.NextStep()
```

### Step 3: Handle Dynamic Contract ABI Input (Optional but Recommended)

**File**: `./evm-events-calls/convo.go`

**Location**: Find the `case InputDynamicContractABI:` handler (around line 373-387)

**Current Code**:
```go
case InputDynamicContractABI:
    factory := c.contextContract()
    if factory == nil {
        return QuitInvalidContext
    }

    contract := c.State.dynamicContractOf(factory.Name)

    rawMessage := json.RawMessage(msg.Value)
    if _, err := json.Marshal(rawMessage); err != nil {
        return loop.Seq(c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskDynamicContractABI{}))
    }

    contract.RawABI = rawMessage
    return c.NextStep()
```

**Change Required**: Add `contract.abiFetchedInThisSession = true` before returning:

```go
case InputDynamicContractABI:
    factory := c.contextContract()
    if factory == nil {
        return QuitInvalidContext
    }

    contract := c.State.dynamicContractOf(factory.Name)

    rawMessage := json.RawMessage(msg.Value)
    if _, err := json.Marshal(rawMessage); err != nil {
        return loop.Seq(c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskDynamicContractABI{}))
    }

    contract.RawABI = rawMessage
    contract.abiFetchedInThisSession = true  // ADD THIS LINE
    return c.NextStep()
```

---

## Verification Steps

### 1. Build Verification
```bash
go build ./...
```
**Expected**: Clean build with no errors or warnings

### 2. Unit Tests
```bash
go test ./evm-events-calls/... -v
```
**Expected**: All existing tests pass

### 3. Full Test Suite
```bash
go test ./... -v
```
**Expected**: All tests in the repository pass

### 4. Manual Verification (Recommended)

Test the conversation flow to verify the ABI preview is shown:

1. Start the codegen for `evm-events-calls`
2. Select a chain without API support (e.g., a custom/local chain) OR simulate manual ABI input
3. Provide a valid ABI via string input
4. **Verify**: The system should display:
   ```
   Ok, here's what the ABI would produce:

   ```protobuf
   // Events
   message EventName {
     string fieldName = 1;
     ...
   }

   // Calls
   message CallName {
     ...
   }
   ```

5. The system should then ask: "Do you want to proceed with this ABI?"

---

## Reference: Working Example (Solana Anchor)

For comparison, the Solana Anchor generator correctly shows the IDL preview in `./sol-anchor/convo.go`:

```go
// Lines 205-221 in sol-anchor/convo.go
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

Note: The Solana implementation ALWAYS shows the preview (no conditional flag check), which is the correct behavior.

---

## Project Standards Checklist
- [ ] Code follows Go conventions and existing patterns in the repository
- [ ] Error messages are descriptive and actionable
- [ ] Changes are minimal and focused on fixing the specific issue
- [ ] All unit tests pass
- [ ] Build completes successfully

---

## Key Files Reference

| File | Purpose | GitHub Link |
|------|---------|-------------|
| `./evm-events-calls/convo.go` | Main conversation handler with ABI input handlers | https://github.com/streamingfast/substreams-codegen/blob/develop/evm-events-calls/convo.go |
| `./evm-events-calls/state.go` | State structs including `abiFetchedInThisSession` field | https://github.com/streamingfast/substreams-codegen/blob/develop/evm-events-calls/state.go |
| `./sol-anchor/convo.go` | Working example of IDL preview (reference) | https://github.com/streamingfast/substreams-codegen/blob/develop/sol-anchor/convo.go |

---

## Summary of Changes

| Handler | Line (approx) | Change |
|---------|---------------|--------|
| `InputContractABIString` | ~327 | Add `contract.abiFetchedInThisSession = true` |
| `InputContractABIFile` | ~361 | Add `contract.abiFetchedInThisSession = true` |
| `InputDynamicContractABI` | ~386 | Add `contract.abiFetchedInThisSession = true` |

Total lines changed: 3 (one line added in each handler)
