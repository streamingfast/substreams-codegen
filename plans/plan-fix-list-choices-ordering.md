# Implementation Plan for Codegen Agent - Comprehensive Fix for Choices Ordering

## Objective
Fix the bug where list choices are rendered in random order due to improper map iteration in the **server-side** code. The server is sending choices in random order to the client, which displays them as received. Submit PR to `github.com/streamingfast/substreams-codegen`.

## Critical Understanding

### The Real Problem
The client in `substreams/cmd/substreams/init.go` already correctly preserves order for generic ListSelect messages (lines 395-407). However, **the server is sending generators in random order** due to map iteration in `registry.go`.

### Root Cause Analysis
In `substreams-codegen/registry.go`, the `ListConversationHandlers()` function:

```go
func ListConversationHandlers() []*ConversationHandler {
    var handlers []*ConversationHandler
    for _, handler := range Registry {  // ← BUG: Map iteration is random
        handlers = append(handlers, handler)
    }
    sort.Slice(handlers, func(i, j int) bool {
        return handlers[i].Weight > handlers[j].Weight // highest weight first
    })

    return handlers
}
```

**The issue**: When multiple generators have the **same weight**, they appear in random order within that weight group because:
1. Map iteration order is non-deterministic in Go
2. The sort only orders by weight (stable sort doesn't help here since we're starting with random order)

**Example**: All three EVM generators have similar weights:
- `evm-hello-world`: weight 83
- `evm-events-calls`: weight 82
- `evm-events-calls-raw`: weight 82

The two generators with weight 82 will randomly swap positions between runs.

---

## Implementation Steps

### Step 1: Fix the ListConversationHandlers Function

**File to modify**: `substreams-codegen/registry.go`

**Current problematic code** (lines 40-50):
```go
func ListConversationHandlers() []*ConversationHandler {
	var handlers []*ConversationHandler
	for _, handler := range Registry {
		handlers = append(handlers, handler)
	}
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].Weight > handlers[j].Weight // heighest weight first
	})

	return handlers
}
```

**Fixed code**:
```go
func ListConversationHandlers() []*ConversationHandler {
	var handlers []*ConversationHandler
	for _, handler := range Registry {
		handlers = append(handlers, handler)
	}
	sort.Slice(handlers, func(i, j int) bool {
		// First sort by weight (highest first)
		if handlers[i].Weight != handlers[j].Weight {
			return handlers[i].Weight > handlers[j].Weight
		}
		// For same weight, sort alphabetically by ID for deterministic ordering
		return handlers[i].ID < handlers[j].ID
	})

	return handlers
}
```

**Key changes**:
1. Add secondary sort key (ID) for generators with the same weight
2. This ensures deterministic ordering even when weights are equal
3. Alphabetical ordering by ID is intuitive and predictable

**Alternative approach** (if you want to preserve insertion order within same weight):
```go
func ListConversationHandlers() []*ConversationHandler {
	var handlers []*ConversationHandler

	// Collect all unique handler IDs and sort them for deterministic iteration
	var ids []string
	for id := range Registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Build handlers slice in sorted ID order
	for _, id := range ids {
		handlers = append(handlers, Registry[id])
	}

	// Sort by weight (and now ties will maintain alphabetical ID order)
	sort.SliceStable(handlers, func(i, j int) bool {
		return handlers[i].Weight > handlers[j].Weight
	})

	return handlers
}
```

**Choose the first approach** (simpler and more maintainable).

---

## Verification Steps

### 1. Build Verification
```bash
# In the substreams-codegen repository
cd substreams-codegen
go build ./cmd/server
```
Expected: Clean build with no errors

### 2. Unit Test the Fix

Add a new test in `registry_test.go` (create if it doesn't exist):

```go
package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListConversationHandlers_DeterministicOrdering(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := Registry
	defer func() { Registry = originalRegistry }()

	// Create a new registry with test data
	Registry = make(map[string]*ConversationHandler)

	// Register handlers with same weights to test secondary sorting
	RegisterConversation("zebra-gen", "Zebra", "Last alphabetically", func() Converser { return nil }, 82, "TestGroup")
	RegisterConversation("alpha-gen", "Alpha", "First alphabetically", func() Converser { return nil }, 82, "TestGroup")
	RegisterConversation("middle-gen", "Middle", "Middle alphabetically", func() Converser { return nil }, 82, "TestGroup")
	RegisterConversation("high-weight", "High Weight", "Highest weight", func() Converser { return nil }, 90, "TestGroup")
	RegisterConversation("low-weight", "Low Weight", "Lowest weight", func() Converser { return nil }, 70, "TestGroup")

	// Run the function multiple times to ensure consistent ordering
	var firstRun []*ConversationHandler
	for run := 0; run < 10; run++ {
		handlers := ListConversationHandlers()

		if run == 0 {
			firstRun = handlers
		} else {
			// Verify the order is identical to first run
			require.Equal(t, len(firstRun), len(handlers), "run %d: handler count mismatch", run)
			for i := range handlers {
				assert.Equal(t, firstRun[i].ID, handlers[i].ID, "run %d: position %d has different handler", run, i)
			}
		}

		// Verify the expected order
		require.Len(t, handlers, 5)
		assert.Equal(t, "high-weight", handlers[0].ID, "highest weight should be first")

		// The three handlers with weight 82 should be in alphabetical order
		assert.Equal(t, "alpha-gen", handlers[1].ID, "weight 82 group should be alphabetically sorted")
		assert.Equal(t, "middle-gen", handlers[2].ID, "weight 82 group should be alphabetically sorted")
		assert.Equal(t, "zebra-gen", handlers[3].ID, "weight 82 group should be alphabetically sorted")

		assert.Equal(t, "low-weight", handlers[4].ID, "lowest weight should be last")
	}
}

func TestListConversationHandlers_WeightOrdering(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := Registry
	defer func() { Registry = originalRegistry }()

	// Create a new registry with test data
	Registry = make(map[string]*ConversationHandler)

	RegisterConversation("gen-50", "Weight 50", "Low weight", func() Converser { return nil }, 50, "TestGroup")
	RegisterConversation("gen-100", "Weight 100", "High weight", func() Converser { return nil }, 100, "TestGroup")
	RegisterConversation("gen-75", "Weight 75", "Mid weight", func() Converser { return nil }, 75, "TestGroup")

	handlers := ListConversationHandlers()

	require.Len(t, handlers, 3)

	// Verify weights are in descending order
	assert.Equal(t, 100, handlers[0].Weight)
	assert.Equal(t, 75, handlers[1].Weight)
	assert.Equal(t, 50, handlers[2].Weight)
}
```

Run the test:
```bash
go test ./... -v -run TestListConversationHandlers
```

Expected: All tests pass

### 3. Manual Integration Testing

Start the server and test with the actual client:

```bash
# Terminal 1: Start the server
cd substreams-codegen
go run ./cmd/server

# Terminal 2: Run substreams init multiple times
cd substreams
./substreams init --state-file /tmp/test1.json
# Select EVM protocol
# Note the order of project types shown

# Cancel and run again
./substreams init --state-file /tmp/test2.json
# Select EVM protocol
# Verify order matches first run

# Repeat 3-5 times to confirm consistency
```

**Expected behavior**:
After selecting EVM protocol, the project type choices should **always** appear in this order:
1. `evm-events-calls` (weight 82, alphabetically first among weight 82)
2. `evm-events-calls-raw` (weight 82, alphabetically second among weight 82)
3. `evm-hello-world` (weight 83)

Wait... that's wrong. Higher weight should come first! So the correct order should be:
1. `evm-hello-world` (weight 83, highest)
2. `evm-events-calls` (weight 82, alphabetically first among weight 82)
3. `evm-events-calls-raw` (weight 82, alphabetically second among weight 82)

### 4. Verify All Weights Are Unique (Optional but Recommended)

You may want to check if we should adjust weights to make them all unique, which would be cleaner:

```bash
cd substreams-codegen
grep -r "RegisterConversation" --include="*.go" | grep -E "8[0-9]," | sort
```

This will show all registrations and their weights. If you find many duplicates, consider updating weights to be unique.

---

## Testing Strategy Summary

1. **Unit tests** ensure the sorting logic works correctly
2. **Manual testing** with multiple runs confirms no randomness
3. **Integration testing** with real client confirms end-to-end behavior

---

## Project Standards Checklist

- [ ] Code follows Go conventions
- [ ] No new dependencies added
- [ ] Fix is minimal and focused (only changes the sort comparison)
- [ ] Tests added to prevent regression
- [ ] Comments explain secondary sort key rationale
- [ ] Existing tests still pass

---

## Commit Message

```
fix: ensure deterministic ordering for generators with same weight

When multiple conversation handlers have the same weight, they were
appearing in random order due to map iteration before sorting.

Added a secondary sort key (handler ID) to ensure deterministic
alphabetical ordering within the same weight group. This provides
a consistent user experience where choices appear in the same order
every time.

Fixes the issue where EVM generators (evm-events-calls and
evm-events-calls-raw, both weight 82) would randomly swap positions.
```

---

## Files to Modify

1. **Primary fix**: `substreams-codegen/registry.go` (modify `ListConversationHandlers` function)
2. **Tests**: `substreams-codegen/registry_test.go` (create or add tests)

---

## Success Criteria

✅ Running `substreams init` 10 times shows identical ordering each time
✅ Unit tests pass
✅ Integration tests with real client show consistent ordering
✅ All existing tests still pass
✅ Code review approved

---

## Notes

- The client-side fix in the `substreams` repository should already be merged (protocol selection ordering)
- This server-side fix complements that by ensuring the data sent to the client is also deterministically ordered
- Together, these fixes ensure complete ordering consistency throughout the entire conversation flow
