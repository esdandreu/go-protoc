package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Mock BinCache implementation for testing
type mockBinCache struct {
	binPath   string
	err       error
	callCount int
	lastTag   string
}

func (m *mockBinCache) BinPath(tag string) (string, error) {
	m.callCount++
	m.lastTag = tag
	if m.err != nil {
		return "", m.err
	}
	return m.binPath, nil
}

// Helper to create a temporary mock binary for testing
func createMockBinary(t testing.TB) string {
	t.Helper()

	tempDir := t.TempDir()
	binName := "mock-protoc"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(tempDir, binName)

	// Create a simple shell script/batch file that echoes its arguments
	var content string
	if runtime.GOOS == "windows" {
		content = "@echo off\necho mock-protoc called with args: %*\n"
	} else {
		content = "#!/bin/sh\necho \"mock-protoc called with args: $@\"\n"
	}

	if err := os.WriteFile(binPath, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create mock binary: %v", err)
	}

	return binPath
}

func TestDefaultProtocTag(t *testing.T) {
	if DefaultProtocTag != "latest" {
		t.Errorf("Expected DefaultProtocTag to be 'latest', got %q", DefaultProtocTag)
	}
}

func TestRunProtoc_Success(t *testing.T) {
	cache := &mockBinCache{
		binPath: createMockBinary(t),
	}
	dirFs := os.DirFS(t.TempDir())

	// Save and restore environment
	originalTag := os.Getenv("PROTOC_RELEASE_TAG")
	defer func() {
		if originalTag == "" {
			os.Unsetenv("PROTOC_RELEASE_TAG")
		} else {
			os.Setenv("PROTOC_RELEASE_TAG", originalTag)
		}
	}()

	// Test with specific tag
	os.Setenv("PROTOC_RELEASE_TAG", "v25.3")

	err := runProtoc(cache, dirFs, "--version")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify cache was called with correct tag
	if cache.callCount != 1 {
		t.Errorf("Expected cache to be called once, got %d calls", cache.callCount)
	}
	if cache.lastTag != "v25.3" {
		t.Errorf("Expected cache to be called with 'v25.3', got %q", cache.lastTag)
	}
}

func TestRunProtoc_DefaultTag(t *testing.T) {
	cache := &mockBinCache{
		binPath: createMockBinary(t),
	}
	dirFs := os.DirFS(t.TempDir())

	// Save and restore environment
	originalTag, isSet := os.LookupEnv("PROTOC_RELEASE_TAG")
	if !isSet {
		defer os.Unsetenv("PROTOC_RELEASE_TAG")
	} else {
		os.Setenv("PROTOC_RELEASE_TAG", originalTag)
	}

	// Unset environment variable to test default
	os.Unsetenv("PROTOC_RELEASE_TAG")

	err := runProtoc(cache, dirFs, "--help")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify cache was called with default tag
	if cache.lastTag != DefaultProtocTag {
		t.Errorf("Expected cache to be called with default tag %q, got %q", DefaultProtocTag, cache.lastTag)
	}
}

func TestRunProtoc_BinCacheError(t *testing.T) {
	expectedErr := errors.New("failed to resolve binary path")
	cache := &mockBinCache{
		err: expectedErr,
	}
	dirFs := os.DirFS(t.TempDir())

	err := runProtoc(cache, dirFs, "--version")
	if err == nil {
		t.Fatal("Expected error from BinCache, got nil")
	}

	// Check error message contains the expected error
	if !strings.Contains(err.Error(), "failed to get protoc binary") {
		t.Errorf("Expected error message to contain 'failed to get protoc binary', got: %v", err)
	}
	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("Expected error message to contain original error %q, got: %v", expectedErr.Error(), err)
	}
}

func TestRunProtoc_CommandExecutionError(t *testing.T) {
	// Use a non-existent binary path to trigger exec error
	cache := &mockBinCache{
		binPath: "/non/existent/binary",
	}
	dirFs := os.DirFS(t.TempDir())

	err := runProtoc(cache, dirFs, "--version")
	if err == nil {
		t.Fatal("Expected error from command execution, got nil")
	}

	// Should be a path error or exec error
	if !strings.Contains(err.Error(), "no such file") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected path/exec error, got: %v", err)
	}
}

func TestRunProtoc_MultipleArguments(t *testing.T) {
	cache := &mockBinCache{
		binPath: createMockBinary(t),
	}
	dirFs := os.DirFS(t.TempDir())

	// Save and restore environment
	originalTag := os.Getenv("PROTOC_RELEASE_TAG")
	defer func() {
		if originalTag == "" {
			os.Unsetenv("PROTOC_RELEASE_TAG")
		} else {
			os.Setenv("PROTOC_RELEASE_TAG", originalTag)
		}
	}()

	os.Setenv("PROTOC_RELEASE_TAG", "v25.3")

	// Test with multiple arguments
	err := runProtoc(cache, dirFs, "--help", "--version")
	if err != nil {
		t.Errorf("Expected no error with multiple args, got: %v", err)
	}
}

func TestRunProtoc_NoArguments(t *testing.T) {
	cache := &mockBinCache{
		binPath: createMockBinary(t),
	}
	dirFs := os.DirFS(t.TempDir())

	// Test with no arguments
	err := runProtoc(cache, dirFs)
	if err != nil {
		t.Errorf("Expected no error with no args, got: %v", err)
	}
}

func TestRunProtoc_EnvironmentVariableHandling(t *testing.T) {
	binPath := createMockBinary(t)
	dirFs := os.DirFS(t.TempDir())

	// Save original environment
	originalTag := os.Getenv("PROTOC_RELEASE_TAG")
	defer func() {
		if originalTag == "" {
			os.Unsetenv("PROTOC_RELEASE_TAG")
		} else {
			os.Setenv("PROTOC_RELEASE_TAG", originalTag)
		}
	}()

	testCases := []struct {
		name        string
		envValue    string
		setEnv      bool
		expectedTag string
	}{
		{
			name:        "Environment variable not set",
			setEnv:      false,
			expectedTag: DefaultProtocTag,
		},
		{
			name:        "Environment variable set to specific version",
			envValue:    "v25.3",
			setEnv:      true,
			expectedTag: "v25.3",
		},
		{
			name:        "Environment variable set to latest",
			envValue:    "latest",
			setEnv:      true,
			expectedTag: "latest",
		},
		{
			name:        "Environment variable set to empty string",
			envValue:    "",
			setEnv:      true,
			expectedTag: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := &mockBinCache{
				binPath: binPath,
			}

			// Set up environment
			if tc.setEnv {
				os.Setenv("PROTOC_RELEASE_TAG", tc.envValue)
			} else {
				os.Unsetenv("PROTOC_RELEASE_TAG")
			}

			err := runProtoc(cache, dirFs, "--version")
			if err != nil {
				t.Errorf("Expected no error for %s, got: %v", tc.name, err)
			}

			if cache.lastTag != tc.expectedTag {
				t.Errorf("Expected tag %q, got %q", tc.expectedTag, cache.lastTag)
			}
		})
	}
}

func TestRunProtoc_BinCacheCalledOnce(t *testing.T) {
	cache := &mockBinCache{
		binPath: createMockBinary(t),
	}
	dirFs := os.DirFS(t.TempDir())

	err := runProtoc(cache, dirFs, "--version")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if cache.callCount != 1 {
		t.Errorf("Expected BinCache.BinPath to be called exactly once, got %d calls", cache.callCount)
	}
}

func TestRunProtoc_ErrorMessageFormat(t *testing.T) {
	expectedErr := errors.New("cache resolution failed")
	cache := &mockBinCache{
		err: expectedErr,
	}
	dirFs := os.DirFS(t.TempDir())

	// Save and restore environment
	originalTag := os.Getenv("PROTOC_RELEASE_TAG")
	defer func() {
		if originalTag == "" {
			os.Unsetenv("PROTOC_RELEASE_TAG")
		} else {
			os.Setenv("PROTOC_RELEASE_TAG", originalTag)
		}
	}()

	testTag := "v25.3"
	os.Setenv("PROTOC_RELEASE_TAG", testTag)

	err := runProtoc(cache, dirFs, "--version")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check that error message includes the tag
	if !strings.Contains(err.Error(), testTag) {
		t.Errorf("Expected error message to contain tag %q, got: %v", testTag, err)
	}

	// Check that error message includes the original error
	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("Expected error message to contain original error %q, got: %v", expectedErr.Error(), err)
	}
}
