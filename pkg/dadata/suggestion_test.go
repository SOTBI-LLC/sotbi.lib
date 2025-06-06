package dadata

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSuggestion(t *testing.T) {
	suggestion := New(os.Getenv("DADATA_API_KEY"))
	suggestions, err := suggestion.Get("7707083893")

	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	require.Equal(t, suggestions[0].Data.Inn, "7707083893")
}
