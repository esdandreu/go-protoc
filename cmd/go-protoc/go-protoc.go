package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"slices"

	"github.com/esdandreu/go-protoc/pkg/bincache"
)

const (
	DefaultProtocTag         = "latest"
	DefaultGoOutFlag         = "--go_out=."
	DefaultGoOptPathFlag     = "--go_opt=paths=source_relative"
	DefaultGoGrpcOutFlag     = "--go-grpc_out=."
	DefaultGoGrpcOptFlag     = "--go-grpc_opt=paths=source_relative"
	DefaultProtoFilesPattern = "**/*.proto"
)

type BinCache interface {
	BinPath(tag string) (string, error)
}

func runProtoc(cache BinCache, dirFs fs.FS, args ...string) error {
	// Determine protoc release tag.
	tag, ok := os.LookupEnv("PROTOC_RELEASE_TAG")
	if !ok {
		tag = DefaultProtocTag
	}

	flags, hasNonFlagArgs := ParseArgs(args)
	if !slices.Contains(flags, "go_out") {
		args = append(args, DefaultGoOutFlag)
	}
	if !slices.Contains(flags, "go_opt") {
		args = append(args, DefaultGoOptPathFlag)
	}
	if !slices.Contains(flags, "go-grpc_out") {
		args = append(args, DefaultGoGrpcOutFlag)
	}
	if !slices.Contains(flags, "go-grpc_opt") {
		args = append(args, DefaultGoGrpcOptFlag)
	}
	if !hasNonFlagArgs {
		matches, err := fs.Glob(dirFs, DefaultProtoFilesPattern)
		if err != nil {
			return fmt.Errorf("failed to glob proto files: %w", err)
		}
		args = append(args, matches...)
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
	dirFs := os.DirFS(".")
	if err := runProtoc(cache, dirFs, os.Args[1:]...); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		log.Fatalf("Failed to run protoc: %v", err)
	}
}
