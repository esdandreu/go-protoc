package bincache

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/esdandreu/go-protoc/pkg/downloader"
	"github.com/esdandreu/go-protoc/pkg/releases"
)

const DefaultProtocBinCachePrefix = "go-protoc"

type VersionResolver interface {
	// ResolveVersion returns the version string for a given tag. As a special
	// case, if the tag is "latest", the latest version should be returned.
	// Otherwise, it sanitizes the input tag into a valid version string
	// without the 'v' prefix.
	ResolveVersion(tag string) (string, error)
}

type URLResolver interface {
	// ResolveURL returns the URL to the protoc binary for a given version,
	// operating system, and architecture.
	ResolveURL(version, goos, goarch string) (*url.URL, error)
}

type ZipDownloader interface {
	DownloadAndExtract(url string, destDir string) error
}

type ProtocBinCache struct {
	VersionResolver
	URLResolver
	ZipDownloader
	path   string
	goos   string
	goarch string
}

// NewProtocBinCache creates a new protoc binary cache. Typically constructed
// with the result of os.UserCacheDir().
func NewProtocBinCache(cacheDir string) *ProtocBinCache {
	return &ProtocBinCache{
		VersionResolver: releases.NewProtocVersionResolver(),
		URLResolver:     releases.NewProtocURLResolver(),
		ZipDownloader:   downloader.NewZipDownloader(),
		path:            path.Join(cacheDir, DefaultProtocBinCachePrefix),
		goos:            runtime.GOOS,
		goarch:          runtime.GOARCH,
	}
}

// BinPath returns the path to the protoc binary in the cache. It will download
// the release if it is not already cached.
func (protoc *ProtocBinCache) BinPath(tag string) (string, error) {
	// Resolve the tag to a version.
	version, err := protoc.ResolveVersion(tag)
	if err != nil {
		return "", fmt.Errorf("failed to resolve version: %w", err)
	}

	// Create version-specific cache directory
	versionDir := filepath.Join(protoc.path, version)
	binPath := filepath.Join(versionDir, "bin", "protoc")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	// Check if binary already exists
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check binary: %w", err)
	}

	// Binary doesn't exist, need to download and extract
	downloadURL, err := protoc.ResolveURL(version, protoc.goos, protoc.goarch)
	if err != nil {
		return "", fmt.Errorf("failed to resolve URL: %w", err)
	}

	// Create the version directory
	if err := os.MkdirAll(versionDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Download and extract the zip file
	if err := protoc.DownloadAndExtract(downloadURL.String(), versionDir); err != nil {
		return "", fmt.Errorf("failed to download and extract: %w", err)
	}

	// Verify the binary now exists
	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("binary not found after extraction: %w", err)
	}

	return binPath, nil
}
