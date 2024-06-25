package remotebuild

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	pbbuild "github.com/streamingfast/substreams-codegen/pb/sf/codegen/remotebuild/v1"
)

func CollectArtifacts(dir string, pattern string) (out []*pbbuild.BuildResponse_BuildArtifact, err error) {
	// Currently pattern will always be *.spkg
	matching, err := doublestar.Glob(os.DirFS(dir), pattern)
	if err != nil {
		return nil, fmt.Errorf("reading output dir: %w", err)
	}

	for _, file := range matching {
		content, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", filepath.Join(dir, file), err)
		}
		out = append(out, &pbbuild.BuildResponse_BuildArtifact{
			Filename: file,
			Content:  content,
		})
	}

	return
}
