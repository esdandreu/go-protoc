package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/esdandreu/go-protoc/pkg/bincache"
)

const defaultProtocVersion = "v32.0"

func main() {
	// Create binary cache
	cache, err := bincache.NewProtocBinCache()
	if err != nil {
		log.Fatalf("Failed to create protoc cache: %v", err)
	}

	// Get protoc binary path (downloads if needed)
	version, ok := os.LookupEnv("PROTOC_RELEASE_TAG")
	if !ok {
		version = defaultProtocVersion
	}

	binPath, err := cache.BinPath(version)
	if err != nil {
		log.Fatalf("Failed to get protoc binary v%s: %v", version, err)
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
