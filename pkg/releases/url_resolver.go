package releases

import "net/url"

type URLResolver interface {
	// ResolveURL returns the URL to the protoc binary for a given version,
	// operating system, and architecture.
	ResolveURL(version, goos, goarch string) (*url.URL, error)
}
