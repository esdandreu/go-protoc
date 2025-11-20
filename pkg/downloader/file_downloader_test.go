package downloader

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFileDownloader_DownloadFile_Success(t *testing.T) {
	content := "Hello, World!"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	downloader := NewFileDownloader()
	var buf bytes.Buffer

	bytesWritten, err := downloader.DownloadFile(server.URL, &buf)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if bytesWritten != int64(len(content)) {
		t.Fatalf("Expected %d bytes written, got %d", len(content), bytesWritten)
	}
	if buf.String() != content {
		t.Fatalf("Expected content %q, got %q", content, buf.String())
	}
}

func TestFileDownloader_DownloadFile_MissingContentLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The HTTP library will treat invalid Content-Length as missing
		// so this will result in a missing Content-Length error
		w.Header().Set("Content-Length", "invalid-number")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("some content"))
	}))
	defer server.Close()

	downloader := NewFileDownloader()
	var buf bytes.Buffer

	bytesWritten, err := downloader.DownloadFile(server.URL, &buf)

	if err == nil {
		t.Fatal("Expected error for invalid Content-Length header")
	}
	if bytesWritten != 0 {
		t.Fatalf("Expected 0 bytes written on error, got %d", bytesWritten)
	}
	// The HTTP library treats invalid Content-Length as missing, so we get the missing error
	if !strings.Contains(err.Error(), "Content-Length header is missing") {
		t.Fatalf("Expected error about missing Content-Length, got: %v", err)
	}
}

func TestFileDownloader_DownloadFile_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	downloader := NewFileDownloader()
	var buf bytes.Buffer

	bytesWritten, err := downloader.DownloadFile(server.URL, &buf)

	if err == nil {
		t.Fatal("Expected error for HTTP 404 status")
	}
	if bytesWritten != 0 {
		t.Fatalf("Expected 0 bytes written on error, got %d", bytesWritten)
	}
	if !strings.Contains(err.Error(), "bad status") {
		t.Fatalf("Expected error about bad status, got: %v", err)
	}
}

func TestFileDownloader_DownloadFile_ContentLengthMismatch(t *testing.T) {
	content := "Hello, World!"
	wrongLength := len(content) + 5 // Claim more bytes than actually provided

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", wrongLength))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	downloader := NewFileDownloader()
	var buf bytes.Buffer

	bytesWritten, err := downloader.DownloadFile(server.URL, &buf)

	if err == nil {
		t.Fatal("Expected error for content length mismatch")
	}
	if bytesWritten != int64(len(content)) {
		t.Fatalf("Expected %d bytes written, got %d", len(content), bytesWritten)
	}
	if !strings.Contains(err.Error(), "expected") && !strings.Contains(err.Error(), "got") {
		t.Fatalf("Expected error about byte count mismatch, got: %v", err)
	}
}

func TestFileDownloader_DownloadFile_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	downloader := NewFileDownloader()
	var buf bytes.Buffer

	bytesWritten, err := downloader.DownloadFile(server.URL, &buf)

	if err != nil {
		t.Fatalf("Expected no error for empty content, got: %v", err)
	}
	if bytesWritten != 0 {
		t.Fatalf("Expected 0 bytes written for empty content, got %d", bytesWritten)
	}
	if buf.Len() != 0 {
		t.Fatalf("Expected empty buffer, got %d bytes", buf.Len())
	}
}

func TestFileDownloader_DownloadFile_WriterError(t *testing.T) {
	content := "Hello, World!"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	// Create a writer that will fail after a few bytes
	errorWriter := &errorAfterNBytesWriter{maxBytes: 5}

	downloader := NewFileDownloader()
	bytesWritten, err := downloader.DownloadFile(server.URL, errorWriter)

	if err == nil {
		t.Fatal("Expected error from failing writer")
	}
	if bytesWritten == 0 {
		t.Fatal("Expected some bytes to be written before error occurred")
	}
}

// errorAfterNBytesWriter is a test helper that fails after writing N bytes
type errorAfterNBytesWriter struct {
	written  int
	maxBytes int
}

func (w *errorAfterNBytesWriter) Write(p []byte) (n int, err error) {
	if w.written >= w.maxBytes {
		return 0, fmt.Errorf("simulated writer error")
	}

	toWrite := len(p)
	if w.written+toWrite > w.maxBytes {
		toWrite = w.maxBytes - w.written
	}

	w.written += toWrite
	return toWrite, nil
}
