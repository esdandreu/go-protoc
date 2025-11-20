package releases

import (
	"fmt"
	"net/url"
	"strings"
)

type ProtocURLResolver struct{}

func NewProtocURLResolver() URLResolver {
	return &ProtocURLResolver{}
}

func (resolver *ProtocURLResolver) ResolveURL(version, goos, goarch string) (*url.URL, error) {
	sanitizedVersion := strings.TrimPrefix(version, "v")
	filename := resolver.getPlatformFilename(sanitizedVersion, goos, goarch)
	url := &url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   fmt.Sprintf("/protocolbuffers/protobuf/releases/download/v%s/%s", sanitizedVersion, filename),
	}
	return url, nil
}

func (resolver *ProtocURLResolver) getPlatformFilename(version, goos, goarch string) string {
	// Special handling for Windows which has different naming conventions
	if goos == "windows" {
		// Windows releases use win32 or win64 as the complete platform identifier
		if goarch == "386" {
			return fmt.Sprintf("protoc-%s-win32.zip", version)
		} else {
			// For amd64, arm64, or any other arch, use win64
			return fmt.Sprintf("protoc-%s-win64.zip", version)
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
	return fmt.Sprintf("protoc-%s-%s-%s.zip", version, platform, arch)
}
