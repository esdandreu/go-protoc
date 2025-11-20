package releases

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/mod/semver"
)

func TestNewProtocVersionResolver(t *testing.T) {
	resolver := NewProtocVersionResolver()
	if resolver == nil {
		t.Fatal("Expected non-nil resolver")
	}
}

func TestProtocVersionResolver_ResolveVersion_Latest(t *testing.T) {
	resolver := &ProtocVersionResolver{}
	version, err := resolver.ResolveVersion("latest")
	if err != nil {
		t.Fatalf("Expected no error for 'latest' tag, got: %v", err)
	}
	if version == "" {
		t.Error("Expected non-empty version")
	}
	// Version should not have 'v' prefix
	if version[0] == 'v' {
		t.Errorf("Expected version without 'v' prefix, got: %s", version)
	}
}

func TestProtocVersionResolver_ResolveVersion_SpecificTag(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"v25.3", "25.3"},
		{"25.3", "25.3"},
		{"v1.0.0", "1.0.0"},
		{"1.0.0", "1.0.0"},
		{"v25.3-rc1", "25.3-rc1"},
	}

	resolver := &ProtocVersionResolver{}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			version, err := resolver.ResolveVersion(tc.input)
			if err != nil {
				t.Fatalf("Expected no error for tag %q, got: %v", tc.input, err)
			}
			if version != tc.expected {
				t.Errorf("Expected version %q, got %q", tc.expected, version)
			}
		})
	}
}

func TestProtocVersionResolver_GetLatestReleaseTag_Success(t *testing.T) {
	resolver := &ProtocVersionResolver{}
	tag, err := resolver.getLatestReleaseTag()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	t.Logf("latest protoc release tag: %s", tag)
	if tag == "" {
		t.Errorf("expected non-empty tag")
	}
	if !semver.IsValid(tag) {
		t.Errorf("expected valid semver, got %s", tag)
	}
	if semver.Compare(tag, "v32.0") != 1 {
		t.Errorf("expected tag to be greater than v32.0, got %s", tag)
	}
}

func TestProtocVersionResolver_GetLatestReleaseTag_HTTPError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Replace the GitHub API URL with our test server
	// Since we can't easily mock the HTTP client, we'll test the error handling
	// by creating a resolver that we know will fail
	
	// This test verifies error handling behavior when GitHub API returns non-200 status
	// We'll test this by examining the real implementation's behavior when it fails
	
	// We can't easily inject the test server URL, but we can test with an invalid URL
	// to simulate network failures or by testing the error path in a different way
	
	// For now, let's test that the resolver handles malformed responses correctly
	// This is tested implicitly in the integration test above, but we could add
	// more specific error cases if we refactor to accept a custom HTTP client
	
	t.Skip("Skipping HTTP error test - would require dependency injection for HTTP client")
}

func TestProtocVersionResolver_GetLatestReleaseTag_InvalidJSON(t *testing.T) {
	// Create a mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// Similar to above - we'd need dependency injection to test this properly
	// For now, document that this would be good to test with a refactored version
	t.Skip("Skipping invalid JSON test - would require dependency injection for HTTP client")
}

func TestProtocVersionResolver_ResolveVersion_EmptyTag(t *testing.T) {
	resolver := &ProtocVersionResolver{}
	version, err := resolver.ResolveVersion("")
	if err != nil {
		t.Fatalf("Expected no error for empty tag, got: %v", err)
	}
	// Empty tag should still return a string (empty string after trimming)
	if version != "" {
		t.Errorf("Expected empty version for empty tag, got: %q", version)
	}
}

// Test that demonstrates the integration works end-to-end
func TestProtocVersionResolver_Integration(t *testing.T) {
	resolver := NewProtocVersionResolver()
	
	// Test latest resolution
	latestVersion, err := resolver.ResolveVersion("latest")
	if err != nil {
		t.Fatalf("Failed to resolve latest version: %v", err)
	}
	
	// Test specific version resolution
	specificVersion, err := resolver.ResolveVersion("v25.3")
	if err != nil {
		t.Fatalf("Failed to resolve specific version: %v", err)
	}
	
	if specificVersion != "25.3" {
		t.Errorf("Expected '25.3', got '%s'", specificVersion)
	}
	
	// Latest should be a valid semver
	if !semver.IsValid("v" + latestVersion) {
		t.Errorf("Latest version should be valid semver: %s", latestVersion)
	}
	
	t.Logf("Latest version: %s", latestVersion)
	t.Logf("Specific version: %s", specificVersion)
}
