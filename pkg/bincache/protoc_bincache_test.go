package bincache

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

// Mock implementations for testing
type mockVersionResolver struct {
	version string
	err     error
}

func (m *mockVersionResolver) ResolveVersion(tag string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.version, nil
}

type mockURLResolver struct {
	url *url.URL
	err error
}

func (m *mockURLResolver) ResolveURL(version, goos, goarch string) (*url.URL, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.url, nil
}

type mockZipDownloader struct {
	err       error
	callCount int
}

func (m *mockZipDownloader) DownloadAndExtract(url string, destDir string) error {
	m.callCount++
	if m.err != nil {
		return m.err
	}

	// Create a mock protoc binary for successful downloads
	binDir := filepath.Join(destDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	binName := "protoc"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	// Create a mock binary file
	return os.WriteFile(binPath, []byte("mock protoc binary"), 0755)
}

func TestProtocBinCache_Path(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewProtocBinCache(tempDir)
	expectedPath := path.Join(tempDir, DefaultProtocBinCachePrefix)
	if cache.path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, cache.path)
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

func TestProtocBinCache_BinPath_Success(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewProtocBinCache(tempDir)
	cache.VersionResolver = &mockVersionResolver{version: "25.3", err: nil}
	cache.URLResolver = &mockURLResolver{
		url: &url.URL{Scheme: "https", Host: "example.com", Path: "/protoc.zip"},
	}
	cache.ZipDownloader = &mockZipDownloader{}
	tag := "v25.3"

	binPath, err := cache.BinPath(tag)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify binary path format
	expectedBinName := "protoc"
	if runtime.GOOS == "windows" {
		expectedBinName += ".exe"
	}
	expectedPath := filepath.Join(tempDir, DefaultProtocBinCachePrefix, "25.3", "bin", expectedBinName)
	if binPath != expectedPath {
		t.Errorf("Expected binary path %q, got %q", expectedPath, binPath)
	}

	// Verify binary exists
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("Expected binary to exist at %q, got error: %v", binPath, err)
	}
}

func TestProtocBinCache_BinPath_CachedBinary(t *testing.T) {
	tempDir := t.TempDir()

	mockDownloader := &mockZipDownloader{}
	cache := NewProtocBinCache(tempDir)
	cache.VersionResolver = &mockVersionResolver{version: "25.3", err: nil}
	cache.URLResolver = &mockURLResolver{
		url: &url.URL{Scheme: "https", Host: "example.com", Path: "/protoc.zip"},
	}
	cache.ZipDownloader = mockDownloader
	tag := "v25.3"

	// First call should download
	binPath1, err := cache.BinPath(tag)
	if err != nil {
		t.Fatalf("Expected no error on first call, got: %v", err)
	}

	// Second call should use cached binary
	binPath2, err := cache.BinPath(tag)
	if err != nil {
		t.Fatalf("Expected no error on second call, got: %v", err)
	}

	if binPath1 != binPath2 {
		t.Errorf("Expected same binary path, got %q and %q", binPath1, binPath2)
	}

	// Download should only be called once
	if mockDownloader.callCount != 1 {
		t.Errorf("Expected download to be called once, got %d calls", mockDownloader.callCount)
	}
}

func TestProtocBinCache_BinPath_VersionResolverError(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewProtocBinCache(tempDir)
	cache.VersionResolver = &mockVersionResolver{err: errors.New("version resolution failed")}
	cache.URLResolver = &mockURLResolver{}
	cache.ZipDownloader = &mockZipDownloader{}
	tag := "invalid-tag"

	_, err := cache.BinPath(tag)
	if err == nil {
		t.Fatal("Expected error for version resolution failure")
	}
	if !errors.Is(err, fmt.Errorf("failed to resolve version: %w", errors.New("version resolution failed"))) {
		// Check error message contains expected text
		expectedMsg := "failed to resolve version"
		if err.Error()[:len(expectedMsg)] != expectedMsg {
			t.Errorf("Expected error to start with %q, got: %v", expectedMsg, err)
		}
	}
}

func TestProtocBinCache_BinPath_URLResolverError(t *testing.T) {
	tempDir := t.TempDir()

	cache := NewProtocBinCache(tempDir)
	cache.VersionResolver = &mockVersionResolver{version: "25.3", err: nil}
	cache.URLResolver = &mockURLResolver{err: errors.New("URL resolution failed")}
	cache.ZipDownloader = &mockZipDownloader{}
	tag := "v25.3"

	_, err := cache.BinPath(tag)
	if err == nil {
		t.Fatal("Expected error for URL resolution failure")
	}
	expectedMsg := "failed to resolve URL"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error to start with %q, got: %v", expectedMsg, err)
	}
}

func TestProtocBinCache_BinPath_DownloadError(t *testing.T) {
	tempDir := t.TempDir()

	cache := NewProtocBinCache(tempDir)
	cache.VersionResolver = &mockVersionResolver{version: "25.3", err: nil}
	cache.URLResolver = &mockURLResolver{
		url: &url.URL{Scheme: "https", Host: "example.com", Path: "/protoc.zip"},
	}
	cache.ZipDownloader = &mockZipDownloader{err: errors.New("download failed")}
	tag := "v25.3"

	_, err := cache.BinPath(tag)
	if err == nil {
		t.Fatal("Expected error for download failure")
	}
	expectedMsg := "failed to download and extract"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error to start with %q, got: %v", expectedMsg, err)
	}
}
