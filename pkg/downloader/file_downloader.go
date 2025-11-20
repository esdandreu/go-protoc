package downloader

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type FileDownloader struct{}

func NewFileDownloader() *FileDownloader {
	return &FileDownloader{}
}

func (downloader *FileDownloader) DownloadFile(url string, w io.Writer) (int64, error) {
	// Get the data.
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Check for HTTP errors.
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status: %s", resp.Status)
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("Content-Length header is missing")
	}
	expectedLength, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Content-Length: %w", err)
	}

	// Write to destination.
	written, err := io.Copy(w, io.LimitReader(resp.Body, expectedLength))
	if err != nil {
		return written, err
	}
	if written != expectedLength {
		return written, fmt.Errorf("expected %d bytes, got %d", expectedLength, written)
	}
	return written, nil
}
