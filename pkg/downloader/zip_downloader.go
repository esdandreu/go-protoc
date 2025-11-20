package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ZipDownloader struct {
	FileDownloader
}

func NewZipDownloader() *ZipDownloader {
	return &ZipDownloader{}
}

// DownloadAndExtract downloads a zip file from the given URL and extracts it
// to the destination directory.
func (downloader *ZipDownloader) DownloadAndExtract(
	url string, destDir string,
) error {
	// Create a temporary file for the download
	tempFile, err := os.CreateTemp("", "protoc-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Download the file
	_, err = downloader.DownloadFile(url, tempFile)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Close the temp file so we can read it
	tempFile.Close()

	// Extract all files from the zip
	return extractAll(tempFile.Name(), destDir)
}

// extractAll extracts all files from a zip archive to the destination
// directory.
func extractAll(zipPath, destDir string) error {
	zipReader, err := zip.OpenReader(zipPath)
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
