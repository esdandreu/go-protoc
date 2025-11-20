package bincache

import (
	"fmt"
	"path"
	"runtime"

	"github.com/esdandreu/go-protoc/pkg/downloader"
	"github.com/esdandreu/go-protoc/pkg/releases"
)

const DefaultProtocBinCachePrefix = "go-protoc"

type protocBinCache struct {
	releases.VersionResolver
	releases.URLResolver
	downloader.FileDownloader
	path   string
	tag    string
	goos   string
	goarch string
}

// ? Should it be able to fail?

// NewProtocBinCache creates a new protoc binary cache. Typically constructed
// with the result of os.UserCacheDir().
func NewProtocBinCache(cacheDir string, tag string) BinCache {
	return &protocBinCache{
		URLResolver:    releases.NewProtocURLResolver(),
		FileDownloader: downloader.NewFileDownloader(),
		path:           path.Join(cacheDir, DefaultProtocBinCachePrefix),
		tag:            tag,
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
	}
}

func (protoc *protocBinCache) Path() string {
	return protoc.path
}

// BinPath returns the path to the protoc binary in the cache. It will download
// the release if it is not already cached.
func (protoc *protocBinCache) BinPath() (string, error) {
	// Resolve the tag to a version.

	//

	return "", fmt.Errorf("not implemented")
}

func (protoc *protocBinCache) Clean() error {
	return nil
}
