package releases

import (
	"sync"
	"testing"
)

func TestNewProtocURLResolver(t *testing.T) {
	resolver := NewProtocURLResolver()
	if resolver == nil {
		t.Fatal("Expected non-nil resolver")
	}
}

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

func TestProtocURLResolver_WindowsPlatforms(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		goarch   string
		expected string
	}{
		{
			name:     "Windows 32-bit",
			version:  "25.3",
			goarch:   "386",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-win32.zip",
		},
		{
			name:     "Windows 64-bit (amd64)",
			version:  "25.3",
			goarch:   "amd64",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-win64.zip",
		},
		{
			name:     "Windows ARM64",
			version:  "25.3",
			goarch:   "arm64",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-win64.zip",
		},
	}

	resolver := &ProtocURLResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := resolver.ResolveURL(tc.version, "windows", tc.goarch)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if url.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, url.String())
			}
		})
	}
}

func TestProtocURLResolver_UnknownPlatforms(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		goos     string
		goarch   string
		expected string
	}{
		{
			name:     "Unknown OS with known arch",
			version:  "25.3",
			goos:     "freebsd",
			goarch:   "amd64",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-freebsd-x86_64.zip",
		},
		{
			name:     "Known OS with unknown arch",
			version:  "25.3",
			goos:     "linux",
			goarch:   "mips64",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-linux-mips64.zip",
		},
		{
			name:     "Both unknown",
			version:  "25.3",
			goos:     "plan9",
			goarch:   "mips",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-plan9-mips.zip",
		},
	}

	resolver := &ProtocURLResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := resolver.ResolveURL(tc.version, tc.goos, tc.goarch)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if url.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, url.String())
			}
		})
	}
}

func TestProtocURLResolver_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		goos     string
		goarch   string
		expected string
	}{
		{
			name:     "Empty version",
			version:  "",
			goos:     "linux",
			goarch:   "amd64",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v/protoc--linux-x86_64.zip",
		},
		{
			name:     "Version with v prefix",
			version:  "v25.3",
			goos:     "linux",
			goarch:   "amd64",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-linux-x86_64.zip",
		},
		{
			name:     "Version without v prefix",
			version:  "25.3",
			goos:     "linux",
			goarch:   "amd64",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-linux-x86_64.zip",
		},
		{
			name:     "Pre-release version",
			version:  "25.3-rc1",
			goos:     "darwin",
			goarch:   "arm64",
			expected: "https://github.com/protocolbuffers/protobuf/releases/download/v25.3-rc1/protoc-25.3-rc1-osx-aarch_64.zip",
		},
	}

	resolver := &ProtocURLResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := resolver.ResolveURL(tc.version, tc.goos, tc.goarch)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if url.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, url.String())
			}
		})
	}
}

func TestProtocURLResolver_GetPlatformFilename(t *testing.T) {
	resolver := &ProtocURLResolver{}

	testCases := []struct {
		name     string
		version  string
		goos     string
		goarch   string
		expected string
	}{
		// Darwin/macOS cases
		{"macOS Intel", "25.3", "darwin", "amd64", "protoc-25.3-osx-x86_64.zip"},
		{"macOS Apple Silicon", "25.3", "darwin", "arm64", "protoc-25.3-osx-aarch_64.zip"},

		// Linux cases
		{"Linux Intel", "25.3", "linux", "amd64", "protoc-25.3-linux-x86_64.zip"},
		{"Linux ARM64", "25.3", "linux", "arm64", "protoc-25.3-linux-aarch_64.zip"},

		// Windows cases (already tested above, but included for completeness)
		{"Windows 32-bit", "25.3", "windows", "386", "protoc-25.3-win32.zip"},
		{"Windows 64-bit", "25.3", "windows", "amd64", "protoc-25.3-win64.zip"},

		// Edge cases
		{"Unknown OS", "25.3", "openbsd", "amd64", "protoc-25.3-openbsd-x86_64.zip"},
		{"Unknown arch", "25.3", "linux", "riscv64", "protoc-25.3-linux-riscv64.zip"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.getPlatformFilename(tc.version, tc.goos, tc.goarch)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}
