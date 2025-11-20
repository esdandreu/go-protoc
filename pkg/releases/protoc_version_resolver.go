package releases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ProtocVersionResolver struct{}

func NewProtocVersionResolver() *ProtocVersionResolver {
	return &ProtocVersionResolver{}
}

func (resolver *ProtocVersionResolver) ResolveVersion(tag string) (string, error) {
	if tag == "latest" {
		var err error
		tag, err = resolver.getLatestReleaseTag()
		if err != nil {
			return "", fmt.Errorf("failed to get latest release tag: %w", err)
		}
	}
	// Version does not have 'v' prefix.
	version := strings.TrimPrefix(tag, "v")

	return version, nil
}

func (resolver *ProtocVersionResolver) getLatestReleaseTag() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/protocolbuffers/protobuf/releases/latest")
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return release.TagName, nil
}
