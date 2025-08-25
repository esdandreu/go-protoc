# go-protoc
protoc wrapper as a go tool for seamless protocol buffers code generation.

## Usage

Use [the `go tool` support available from Go
1.24+](https://tip.golang.org/doc/go1.24#tools) for managing the dependency of
`go-protoc` alongside your core application.

To do this, you run `go get -tool`:

```shell
go get -tool github.com/esdandreu/go-protoc/cmd/go-protoc@latest
```

From there, each invocation of `go-protoc` would be used like so:

```go
//go:generate go tool go-protoc
```

### Usage with go prior to 1.24

If you don't have access to [the `go tool` support available from Go
1.24+](https://tip.golang.org/doc/go1.24#tools), it is recommended to follow
[the `tools.go`
pattern](https://www.jvt.me/posts/2022/06/15/go-tools-dependency-management/)
to still mange the dependency of `go-protoc` alongside your core application.

This would give you a `tools/tools.go`:

```go
//go:build tools
// +build tools

package main

import (
	_ "github.com/esdandreu/go-protoc"
)
```

Then, each invocation of `go-protoc` would be used like so:

```go
//go:generate go run github.com/esdandreu/go-protoc
```

Alternatively, you can install it as a binary with:

```shell
go install github.com/esdandreu/go-protoc@latest
go-protoc -version
```

Which then means you can invoke it like so:

```go
//go:generate go-protoc
```

Note that you can also [move your `tools.go` into its own
sub-module](https://www.jvt.me/posts/2024/09/30/go-tools-module/) to reduce the
impact on your top-level `go.mod`.
