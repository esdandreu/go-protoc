package bincache

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/esdandreu/go-protoc/pkg/releases"
)

type ProtocBinCache struct {
	cacheDir   string
	downloader releases.Downloader
}

// NewProtocBinCache creates a new filesystem-based binary cache
func NewProtocBinCache() (BinCache, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user cache directory: %w", err)
	}

	// Create application-specific cache directory
	cacheDir := filepath.Join(userCacheDir, "go-protoc")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &ProtocBinCache{
		cacheDir:   cacheDir,
		downloader: releases.NewGitHubDownloader(),
	}, nil
}

// Protoc returns the path to the protoc binary for the specified version
// Downloads it if not already cached
func (fc *ProtocBinCache) BinPath(tag string) (string, error) {
	versionCacheDir := filepath.Join(fc.cacheDir, "protoc-"+tag)
	binaryPath := filepath.Join(versionCacheDir, "bin", "protoc")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	// Check if binary already exists in cache
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	// Download and extract protoc binary for the specified version
	if _, err := fc.downloader.Download(
		versionCacheDir, tag, runtime.GOOS, runtime.GOARCH,
	); err != nil {
		return "", fmt.Errorf(
			"failed to download protoc version %s: %w", tag, err,
		)
	}

	// Check if the binary exists after download
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf(
			"protoc binary not found after download: %w", err,
		)
	}

	return binaryPath, nil
}

// CacheDir returns the cache directory path
func (fc *ProtocBinCache) CacheDir() string {
	return fc.cacheDir
}

// Clean removes all cached binaries
func (fc *ProtocBinCache) Clean() error {
	return os.RemoveAll(fc.cacheDir)
}
