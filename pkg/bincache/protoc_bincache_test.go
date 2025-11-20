package bincache

import (
	"errors"
	"os"
	"path"
	"testing"
)

func TestProtocBinCache_Path(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewProtocBinCache(tempDir, "unused")
	expectedPath := path.Join(tempDir, DefaultProtocBinCachePrefix)
	if cache.Path() != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, cache.Path())
	}
	// Creating the cache does not create any directory.
	_, err := os.Stat(expectedPath)
	if err == nil {
		t.Errorf("expected directory %q to not exist", expectedPath)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected directory %q to not exist, got %v", expectedPath, err)
	}
}
