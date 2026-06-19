package main

import (
	"runtime/debug"

	"github.com/uAliAmer/cvgen/cmd"
)

// version is injected at release time via -ldflags by goreleaser.
var version = "dev"

func main() {
	cmd.SetVersion(resolveVersion())
	cmd.Execute()
}

// resolveVersion prefers the ldflags-injected version, then falls back to the
// module version recorded by the Go toolchain (set for `go install pkg@vX`),
// so installed builds report something better than "dev".
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}
