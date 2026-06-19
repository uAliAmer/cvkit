package cv

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// update regenerates the golden file instead of asserting against it:
//
//	go test ./internal/cv -run TestRenderLaTeXGolden -update
var update = flag.Bool("update", false, "update golden files")

// TestRenderLaTeXGolden locks the LaTeX output against a committed golden file.
// The golden was verified byte-identical to the original build_cv.py generator
// (ignoring the generator-name comment), so this guards the parity invariant.
func TestRenderLaTeXGolden(t *testing.T) {
	in := filepath.Join("testdata", "sample.json")
	golden := filepath.Join("testdata", "sample.golden.tex")

	c, err := Load(in)
	if err != nil {
		t.Fatal(err)
	}
	got := c.RenderLaTeX()

	if *update {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("updated %s", golden)
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("RenderLaTeX output drifted from %s.\n"+
			"Run `go test ./internal/cv -run TestRenderLaTeXGolden -update` if intentional.", golden)
	}
}
