package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstreamsSinkChoice_SourceOnly(t *testing.T) {
	t.Helper()

	// Test that source_only enum value exists and is valid
	sourceOnly := SubstreamsSinkChoiceSourceOnly
	assert.Equal(t, "source_only", string(sourceOnly))
	assert.True(t, sourceOnly.IsValid())
}

func TestParseSubstreamsSinkChoice_SourceOnly(t *testing.T) {
	t.Helper()

	parsed, err := ParseSubstreamsSinkChoice("source_only")
	require.NoError(t, err)
	assert.Equal(t, SubstreamsSinkChoiceSourceOnly, parsed)
}

func TestGenerateSinkInstructions_SourceOnly(t *testing.T) {
	t.Helper()

	// Create a test conversation with source-only sink choice
	conv := &Conversation[*BaseConversationState]{
		State: &BaseConversationState{},
	}

	sourceOnly := SubstreamsSinkChoiceSourceOnly
	conv.State.SetSubstreamsSinkChoice(sourceOnly)

	// Should return nil (no sink instructions for source-only)
	instructions := conv.generateSinkInstructions()
	assert.Nil(t, instructions, "source-only should not generate sink instructions")
}

func TestGenerateDockerCompose_SourceOnly(t *testing.T) {
	t.Helper()

	// Create a test conversation with source-only sink choice
	conv := &Conversation[*BaseConversationState]{
		State: &BaseConversationState{},
	}

	sourceOnly := SubstreamsSinkChoiceSourceOnly
	conv.State.SetSubstreamsSinkChoice(sourceOnly)

	// Should return nil (no docker-compose for source-only)
	dockerCompose := conv.generateDockerComposeContent()
	assert.Nil(t, dockerCompose, "source-only should not generate docker-compose")
}

func TestSubstreamsSinkChoices_ContainsSourceOnly(t *testing.T) {
	t.Helper()

	// Verify source-only is in the choices list
	found := false
	for _, choice := range substreamsSinkChoices {
		if choice.Key == "source_only" {
			found = true
			assert.Equal(t, "Just generate the source code (no sink setup)", choice.Value)
			break
		}
	}

	assert.True(t, found, "source_only should be in substreamsSinkChoices")
}

func TestSubstreamsSinkChoices_SourceOnlyIsFirst(t *testing.T) {
	t.Helper()

	// Verify source-only is the first choice (for prominence)
	require.Greater(t, len(substreamsSinkChoices), 0, "substreamsSinkChoices should not be empty")
	assert.Equal(t, "source_only", substreamsSinkChoices[0].Key, "source_only should be the first choice")
}
