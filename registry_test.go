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

func TestListConversationHandlers_RealWorldEVMScenario(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := Registry
	defer func() { Registry = originalRegistry }()

	// Create a new registry with test data mimicking real EVM generators
	Registry = make(map[string]*ConversationHandler)

	// Register handlers similar to actual EVM generators
	RegisterConversation("evm-hello-world", "EVM Hello World", "Hello World", func() Converser { return nil }, 83, "EVM")
	RegisterConversation("evm-events-calls", "EVM Events Calls", "Events Calls", func() Converser { return nil }, 82, "EVM")
	RegisterConversation("evm-events-calls-raw", "EVM Events Calls Raw", "Events Calls Raw", func() Converser { return nil }, 82, "EVM")

	// Run multiple times to ensure consistent ordering
	for run := 0; run < 5; run++ {
		handlers := ListConversationHandlers()

		require.Len(t, handlers, 3)

		// Verify the expected order: highest weight first, then alphabetical within same weight
		assert.Equal(t, "evm-hello-world", handlers[0].ID, "run %d: highest weight (83) should be first", run)
		assert.Equal(t, "evm-events-calls", handlers[1].ID, "run %d: weight 82, alphabetically first", run)
		assert.Equal(t, "evm-events-calls-raw", handlers[2].ID, "run %d: weight 82, alphabetically second", run)
	}
}
