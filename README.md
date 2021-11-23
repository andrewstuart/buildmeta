# Buildmeta

An easy way to set `ldflags` and retain build-time metadata that is useful for
Go applications running in any environment.

## Getting Started

There are already a few applications out there with canonical paths for build
metadata, though none thus far is authoritative. This package takes the idea and
extends it to include a CLI that knows how to set ldflags appropriately so that
all your build needs is `-ldflags "$(buildmeta ldflags)"`. There are additional
options for outputting JSON metadata for e.g. static directories.

To use this, run `go install github.com/andrewstuart/buildmeta/cmd/buildmeta@latest` in
terminal, and run `buildmeta` to validate output. You can also try `go run
github.com/andrewstuart/buildmeta/cmd/buildmeta` which will both download and run this module.

```bash
go build -o app -ldflags "$(buildmeta ldflags)"
```
