package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvkit/internal/cv"
)

var buildAll bool

var buildCmd = &cobra.Command{
	Use:   "build [input] [output]",
	Short: "Render a CV JSON to a LaTeX .tex file",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if buildAll {
			return buildEvery(argOrDefault(args, 0, "."))
		}
		in := argOrDefault(args, 0, "cv_data.json")
		out := argOrDefault(args, 1, deriveTexName(in))
		return buildOne(in, out)
	},
}

func init() {
	buildCmd.Flags().BoolVar(&buildAll, "all", false, "build every cv_data*.json variant in parallel")
	rootCmd.AddCommand(buildCmd)
}

// buildEvery builds all cv_data*.json variants in dir concurrently. One
// variant failing does not abort the others (V6); the first error is returned
// after all finish.
func buildEvery(dir string) error {
	matches, err := filepath.Glob(filepath.Join(dir, "cv_data*.json"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no cv_data*.json files in %s", dir)
	}
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errs   []string
		failed int
	)
	for _, in := range matches {
		wg.Add(1)
		go func(in string) {
			defer wg.Done()
			if err := buildOne(in, deriveTexName(in)); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("%s: %v", in, err))
				failed++
				mu.Unlock()
			}
		}(in)
	}
	wg.Wait()
	fmt.Printf("built %d/%d variants\n", len(matches)-failed, len(matches))
	if len(errs) > 0 {
		return fmt.Errorf("%d failed:\n  %s", failed, strings.Join(errs, "\n  "))
	}
	return nil
}

// buildOne renders a single JSON file to a .tex file.
func buildOne(in, out string) error {
	c, err := cv.Load(in)
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, []byte(c.RenderLaTeX()), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s  ✓   compile: xelatex %s\n", out, filepath.Base(out))
	return nil
}

// deriveTexName maps cv_data.json -> cv.tex and cv_data_<role>.json -> cv_<role>.tex.
func deriveTexName(in string) string {
	return deriveName(in, ".tex")
}

func argOrDefault(args []string, i int, def string) string {
	if i < len(args) && args[i] != "" {
		return args[i]
	}
	return def
}
