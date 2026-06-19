package main

import "github.com/uAliAmer/cvgen/cmd"

// version is injected at release time via -ldflags by goreleaser.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
