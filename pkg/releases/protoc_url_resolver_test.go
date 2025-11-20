package releases

import (
	"sync"
	"testing"
)

func TestProtocURLResolver_ResolveURL(t *testing.T) {
	testCases := map[string]struct {
		version string
		goos    string
		goarch  string
	}{
		// Main release platforms.
		"https://github.com/protocolbuffers/protobuf/releases/download/v32.1/protoc-32.1-linux-aarch_64.zip": {version: "v32.1", goos: "linux", goarch: "arm64"},
		"https://github.com/protocolbuffers/protobuf/releases/download/v32.1/protoc-32.1-linux-x86_64.zip":   {version: "v32.1", goos: "linux", goarch: "amd64"},
		"https://github.com/protocolbuffers/protobuf/releases/download/v32.1/protoc-32.1-osx-aarch_64.zip":   {version: "v32.1", goos: "darwin", goarch: "arm64"},
		"https://github.com/protocolbuffers/protobuf/releases/download/v32.1/protoc-32.1-osx-x86_64.zip":     {version: "v32.1", goos: "darwin", goarch: "amd64"},
		"https://github.com/protocolbuffers/protobuf/releases/download/v32.1/protoc-32.1-win64.zip":          {version: "v32.1", goos: "windows", goarch: "amd64"},
		// v prefix in version is optional.
		"https://github.com/protocolbuffers/protobuf/releases/download/v32.0/protoc-32.0-linux-aarch_64.zip": {version: "32.0", goos: "linux", goarch: "arm64"},
	}

	resolver := &ProtocURLResolver{}
	var wg sync.WaitGroup // Speed up test by running the cases in parallel.
	for expectedURL, args := range testCases {
		wg.Go(func() {
			url, err := resolver.ResolveURL(args.version, args.goos, args.goarch)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if expectedURL != url.String() {
				t.Errorf("expected\n%s\ngot\n%s", expectedURL, url.String())
			}
		})
	}
	wg.Wait()
}
