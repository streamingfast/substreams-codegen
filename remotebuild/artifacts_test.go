package remotebuild_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/streamingfast/substreams-codegen/remotebuild"
	"github.com/stretchr/testify/require"
)

func Test_CollectArtifacts(t *testing.T) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "remotebuild")
	require.NoError(t, err)

	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	tests := []struct {
		name              string
		spkgName          string
		dir               string
		pattern           string
		expectedArtifacts int
	}{
		{
			name:              "Valid spkg file name",
			spkgName:          "substreams.spkg",
			dir:               tempDir,
			pattern:           "*.spkg",
			expectedArtifacts: 1,
		},
		{
			name:              "Another valid spkg file name",
			spkgName:          "./test-v0.1.0.spkg",
			dir:               tempDir,
			pattern:           "*.spkg",
			expectedArtifacts: 1,
		},
		{
			name:              "Invalid spkg file name",
			spkgName:          "test.txt",
			dir:               tempDir,
			pattern:           "*.spkg",
			expectedArtifacts: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spkgFileLocation := filepath.Join(tempDir, test.spkgName)
			err := os.WriteFile(spkgFileLocation, []byte("content"), 0644)
			require.NoError(t, err)

			artifacts, err := remotebuild.CollectArtifacts(tempDir, test.pattern)
			require.NoError(t, err)
			require.Len(t, artifacts, test.expectedArtifacts)

			err = os.Remove(spkgFileLocation)
			require.NoError(t, err)
		})
	}
}
