package bincache

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/esdandreu/go-protoc/pkg/downloader"
	"github.com/esdandreu/go-protoc/pkg/releases"
)

const DefaultProtocBinCachePrefix = "go-protoc"

type protocBinCache struct {
	releases.VersionResolver
	releases.URLResolver
	downloader.FileDownloader
	path   string
	tag    string
	goos   string
	goarch string
}

// ? Should it be able to fail?

// NewProtocBinCache creates a new protoc binary cache. Typically constructed
// with the result of os.UserCacheDir().
func NewProtocBinCache(cacheDir string, tag string) BinCache {
	return &protocBinCache{
		VersionResolver: releases.NewProtocVersionResolver(),
		URLResolver:     releases.NewProtocURLResolver(),
		FileDownloader:  downloader.NewFileDownloader(),
		path:            path.Join(cacheDir, DefaultProtocBinCachePrefix),
		tag:             tag,
		goos:            runtime.GOOS,
		goarch:          runtime.GOARCH,
	}
}

func (protoc *protocBinCache) Path() string {
	return protoc.path
}

// BinPath returns the path to the protoc binary in the cache. It will download
// the release if it is not already cached.
func (protoc *protocBinCache) BinPath() (string, error) {
	// Resolve the tag to a version.
	version, err := protoc.ResolveVersion(protoc.tag)
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
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Download and extract the zip file
	if err := protoc.downloadAndExtract(downloadURL.String(), versionDir); err != nil {
		return "", fmt.Errorf("failed to download and extract: %w", err)
	}

	// Verify the binary now exists
	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("binary not found after extraction: %w", err)
	}

	return binPath, nil
}

// downloadAndExtract downloads a zip file from the given URL and extracts it to the destination directory
func (protoc *protocBinCache) downloadAndExtract(url, destDir string) error {
	// Create a temporary file for the download
	tempFile, err := os.CreateTemp("", "protoc-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Download the file
	_, err = protoc.DownloadFile(url, tempFile)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Reopen the temp file for reading
	tempFile.Close()
	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()

	// Extract files
	for _, file := range zipReader.File {
		destPath := filepath.Join(destDir, file.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, file.FileInfo().Mode())
			continue
		}

		// Create the directories for the file
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract the file
		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			fileReader.Close()
			return fmt.Errorf("failed to create destination file: %w", err)
		}

		_, err = io.Copy(destFile, fileReader)
		fileReader.Close()
		destFile.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	return nil
}

func (protoc *protocBinCache) Clean() error {
	return nil
}
