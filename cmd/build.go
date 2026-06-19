package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvgen/internal/cv"
)

var buildCmd = &cobra.Command{
	Use:   "build [input] [output]",
	Short: "Render a CV JSON to a LaTeX .tex file",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := argOrDefault(args, 0, "cv_data.json")
		out := argOrDefault(args, 1, deriveTexName(in))
		return buildOne(in, out)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
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
	base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
	switch {
	case base == "cv_data":
		base = "cv"
	case strings.HasPrefix(base, "cv_data_"):
		base = "cv_" + strings.TrimPrefix(base, "cv_data_")
	}
	return filepath.Join(filepath.Dir(in), base+".tex")
}

func argOrDefault(args []string, i int, def string) string {
	if i < len(args) && args[i] != "" {
		return args[i]
	}
	return def
}
