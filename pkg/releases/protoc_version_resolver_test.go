package releases

import (
	"testing"

	"golang.org/x/mod/semver"
)

func TestProtocVersionResolver_GetLatestReleaseTag(t *testing.T) {
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
