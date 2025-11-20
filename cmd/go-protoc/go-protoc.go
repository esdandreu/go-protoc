package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/esdandreu/go-protoc/pkg/bincache"
)

const DefaultProtocTag = "latest"

type BinCache interface {
	BinPath(tag string) (string, error)
}

func runProtoc(cache BinCache, args ...string) error {
	// Determine protoc release tag.
	tag, ok := os.LookupEnv("PROTOC_RELEASE_TAG")
	if !ok {
		tag = DefaultProtocTag
	}

	// Get protoc binary path (downloads if needed).
	binPath, err := cache.BinPath(tag)
	if err != nil {
		return fmt.Errorf("failed to get protoc binary %s: %w", tag, err)
	}

	cmd := exec.Command(binPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func main() {
	// Create binary cache
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("failed to get user cache dir: %v", err)
	}
	cache := bincache.NewProtocBinCache(cacheDir)
	if err := runProtoc(cache, os.Args[1:]...); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		log.Fatalf("Failed to run protoc: %v", err)
	}
}
