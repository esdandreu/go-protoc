package releases

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ProtocDownloader implements Downloader for GitHub protobuf releases
type ProtocDownloader struct{}

// NewGitHubDownloader creates a new GitHub releases downloader
func NewGitHubDownloader() *ProtocDownloader {
	return &ProtocDownloader{}
}

// Download downloads and extracts the specified protoc version
func (downloader *ProtocDownloader) Download(
	destDir, version, goos, goarch string,
) (string, error) {
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Get download URL for the specified platform
	downloadURL := downloader.GetDownloadURL(version, goos, goarch)

	// Download the archive
	zipPath := filepath.Join(destDir, "protoc.zip")
	if err := downloader.downloadFile(downloadURL, zipPath); err != nil {
		return "", fmt.Errorf("failed to download protoc archive: %w", err)
	}

	// Extract protoc binary from the archive
	protocPath, err := downloader.extractProtocFromZip(zipPath, destDir, goos)
	if err != nil {
		return "", fmt.Errorf("failed to extract protoc binary: %w", err)
	}

	// Clean up the zip file
	os.Remove(zipPath)

	return protocPath, nil
}

func (downloader *ProtocDownloader) GetDownloadURL(tag, goos, goarch string) string {
	// Ensure tag has 'v' prefix for the download URL
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	filename := downloader.getPlatformFilename(tag, goos, goarch)
	return fmt.Sprintf(
		"https://github.com/protocolbuffers/protobuf/releases/download/%s/%s",
		tag, filename,
	)
}

// getPlatformFilename returns the platform-specific filename for the protoc release
func (downloader *ProtocDownloader) getPlatformFilename(version, goos, goarch string) string {
	// Special handling for Windows which has different naming conventions
	if goos == "windows" {
		// Windows releases use win32 or win64 as the complete platform identifier
		if goarch == "386" {
			return fmt.Sprintf("protoc-%s-win32.zip", strings.TrimPrefix(version, "v"))
		} else {
			// For amd64, arm64, or any other arch, use win64
			return fmt.Sprintf("protoc-%s-win64.zip", strings.TrimPrefix(version, "v"))
		}
	}

	// Map Go's GOOS to protobuf's platform naming
	var platform string
	switch goos {
	case "darwin":
		platform = "osx"
	case "linux":
		platform = "linux"
	default:
		platform = goos // fallback
	}

	// Map Go's GOARCH to protobuf's architecture naming
	var arch string
	switch goarch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch_64"
	default:
		arch = goarch // fallback
	}

	// Construct filename for non-Windows platforms
	return fmt.Sprintf("protoc-%s-%s-%s.zip", strings.TrimPrefix(version, "v"), platform, arch)
}

// downloadFile downloads a file from the given URL to the specified path
func (downloader *ProtocDownloader) downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// extractProtocFromZip extracts the protoc binary from a ZIP archive
func (downloader *ProtocDownloader) extractProtocFromZip(zipPath, extractDir, goos string) (string, error) {
	// Open the ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// Extract files
	for _, file := range reader.File {
		// We're looking for the protoc binary in the bin/ directory
		if strings.HasSuffix(file.Name, "bin/protoc") || strings.HasSuffix(file.Name, "bin/protoc.exe") {
			// Create the bin directory
			binDir := filepath.Join(extractDir, "bin")
			if err := os.MkdirAll(binDir, os.ModePerm); err != nil {
				return "", fmt.Errorf("failed to create bin directory: %w", err)
			}

			// Determine output path
			binaryName := "protoc"
			if goos == "windows" {
				binaryName += ".exe"
			}
			outputPath := filepath.Join(binDir, binaryName)

			// Extract the file
			if err := downloader.extractFile(file, outputPath); err != nil {
				return "", fmt.Errorf("failed to extract protoc binary: %w", err)
			}

			// Make the binary executable (Unix-like systems)
			if goos != "windows" {
				if err := os.Chmod(outputPath, os.ModePerm); err != nil {
					return "", fmt.Errorf("failed to make protoc executable: %w", err)
				}
			}

			return outputPath, nil
		}
	}

	return "", fmt.Errorf("protoc binary not found in archive")
}

// extractFile extracts a single file from a ZIP archive
func (downloader *ProtocDownloader) extractFile(file *zip.File, outputPath string) error {
	// Open the file in the ZIP
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create the output file
	outFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copy the file contents
	_, err = io.Copy(outFile, rc)
	return err
}
