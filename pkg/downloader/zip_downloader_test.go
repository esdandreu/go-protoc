package downloader

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZipDownloader_DownloadAndExtract_Success(t *testing.T) {
	// Create a test zip file in memory
	zipContent := createTestZip(t, map[string]string{
		"test.txt":           "Hello, World!",
		"subdir/nested.txt":  "Nested content",
		"bin/protoc":         "mock protoc binary",
		"bin/protoc.exe":     "mock protoc binary for windows",
	})

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(zipContent)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	downloader := NewZipDownloader()

	err := downloader.DownloadAndExtract(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify extracted files
	testCases := map[string]string{
		"test.txt":           "Hello, World!",
		"subdir/nested.txt":  "Nested content",
		"bin/protoc":         "mock protoc binary",
		"bin/protoc.exe":     "mock protoc binary for windows",
	}

	for filename, expectedContent := range testCases {
		fullPath := filepath.Join(tempDir, filename)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read extracted file %q: %v", filename, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("File %q: expected content %q, got %q", filename, expectedContent, string(content))
		}
	}
}

func TestZipDownloader_DownloadAndExtract_WithDirectories(t *testing.T) {
	// Create a test zip with directories
	zipContent := createTestZipWithDirs(t, map[string]string{
		"dir1/":             "", // directory
		"dir1/file1.txt":    "Content 1",
		"dir2/subdir/":      "", // nested directory
		"dir2/subdir/file2.txt": "Content 2",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(zipContent)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	downloader := NewZipDownloader()

	err := downloader.DownloadAndExtract(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify directory structure
	dirTests := []string{
		"dir1",
		"dir2",
		"dir2/subdir",
	}

	for _, dir := range dirTests {
		fullPath := filepath.Join(tempDir, dir)
		if stat, err := os.Stat(fullPath); err != nil {
			t.Errorf("Directory %q should exist: %v", dir, err)
		} else if !stat.IsDir() {
			t.Errorf("Path %q should be a directory", dir)
		}
	}

	// Verify files
	fileTests := map[string]string{
		"dir1/file1.txt":        "Content 1",
		"dir2/subdir/file2.txt": "Content 2",
	}

	for filename, expectedContent := range fileTests {
		fullPath := filepath.Join(tempDir, filename)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read file %q: %v", filename, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("File %q: expected %q, got %q", filename, expectedContent, string(content))
		}
	}
}

func TestZipDownloader_DownloadAndExtract_DownloadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	downloader := NewZipDownloader()

	err := downloader.DownloadAndExtract(server.URL, tempDir)
	if err == nil {
		t.Fatal("Expected error for HTTP 404")
	}
	if !strings.Contains(err.Error(), "failed to download file") {
		t.Errorf("Expected error about download failure, got: %v", err)
	}
}

func TestZipDownloader_DownloadAndExtract_InvalidZip(t *testing.T) {
	// Create server returning invalid zip content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		invalidZipContent := "this is not a zip file"
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(invalidZipContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(invalidZipContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	downloader := NewZipDownloader()

	err := downloader.DownloadAndExtract(server.URL, tempDir)
	if err == nil {
		t.Fatal("Expected error for invalid zip content")
	}
	if !strings.Contains(err.Error(), "failed to open zip file") {
		t.Errorf("Expected error about zip file, got: %v", err)
	}
}

func TestZipDownloader_DownloadAndExtract_ZipSlipProtection(t *testing.T) {
	// Create a malicious zip with path traversal
	zipContent := createMaliciousZip(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(zipContent)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	downloader := NewZipDownloader()

	err := downloader.DownloadAndExtract(server.URL, tempDir)
	if err == nil {
		t.Fatal("Expected error for zip slip attempt")
	}
	if !strings.Contains(err.Error(), "invalid file path") {
		t.Errorf("Expected error about invalid file path, got: %v", err)
	}
}

func TestZipDownloader_DownloadAndExtract_EmptyZip(t *testing.T) {
	// Create an empty zip file
	zipContent := createTestZip(t, map[string]string{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(zipContent)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	downloader := NewZipDownloader()

	err := downloader.DownloadAndExtract(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Expected no error for empty zip, got: %v", err)
	}

	// Verify destination directory exists but is empty (except for any subdirs created by extraction)
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read destination directory: %v", err)
	}
	
	// Empty zip should result in empty directory
	if len(entries) != 0 {
		t.Errorf("Expected empty directory, found %d entries", len(entries))
	}
}

// Helper function to create a test zip file in memory
func createTestZip(t *testing.T, files map[string]string) []byte {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			t.Fatalf("Failed to create zip entry %q: %v", filename, err)
		}
		_, err = writer.Write([]byte(content))
		if err != nil {
			t.Fatalf("Failed to write zip entry %q: %v", filename, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

// Helper function to create a test zip file with directories
func createTestZipWithDirs(t *testing.T, entries map[string]string) []byte {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for name, content := range entries {
		if strings.HasSuffix(name, "/") {
			// Directory entry - create with proper permissions
			header := &zip.FileHeader{
				Name: name,
				Method: zip.Store,
			}
			header.SetMode(0755 | os.ModeDir)
			_, err := zipWriter.CreateHeader(header)
			if err != nil {
				t.Fatalf("Failed to create directory entry %q: %v", name, err)
			}
		} else {
			// File entry
			writer, err := zipWriter.Create(name)
			if err != nil {
				t.Fatalf("Failed to create zip entry %q: %v", name, err)
			}
			_, err = writer.Write([]byte(content))
			if err != nil {
				t.Fatalf("Failed to write zip entry %q: %v", name, err)
			}
		}
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

// Helper function to create a malicious zip with path traversal
func createMaliciousZip(t *testing.T) []byte {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Create a zip entry that tries to escape the destination directory
	maliciousPath := "../../../malicious.txt"
	writer, err := zipWriter.Create(maliciousPath)
	if err != nil {
		t.Fatalf("Failed to create malicious zip entry: %v", err)
	}
	_, err = writer.Write([]byte("malicious content"))
	if err != nil {
		t.Fatalf("Failed to write malicious zip entry: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close malicious zip writer: %v", err)
	}

	return buf.Bytes()
}
