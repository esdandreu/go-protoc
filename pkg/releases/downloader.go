package releases

// Downloader handles downloading and extracting protoc releases from GitHub
type Downloader interface {
	// Download downloads the specified protoc version to the given directory
	// for the specified OS and architecture. Returns the path to the extracted
	// protoc binary.
	Download(destDir, version, goos, goarch string) (string, error)

	// GetDownloadURL returns the GitHub release download URL for the specified
	// version, OS, and architecture.
	GetDownloadURL(version, goos, goarch string) string
}
