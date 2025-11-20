package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/esdandreu/go-protoc/pkg/bincache"
)

const DefaultProtocTag = "latest"

func main() {
	// Determine protoc release tag.
	tag, ok := os.LookupEnv("PROTOC_RELEASE_TAG")
	if !ok {
		tag = DefaultProtocTag
	}

	// Create binary cache
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("Failed to get user cache dir: %v", err)
	}
	cache := bincache.NewProtocBinCache(cacheDir, tag)
	if err != nil {
		log.Fatalf("Failed to create protoc cache: %v", err)
	}

	// Get protoc binary path (downloads if needed)
	binPath, err := cache.BinPath()
	if err != nil {
		log.Fatalf("Failed to get protoc binary %s: %v", tag, err)
	}

	cmd := exec.Command(binPath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		log.Fatalf("Failed to run protoc: %v", err)
	}
}
