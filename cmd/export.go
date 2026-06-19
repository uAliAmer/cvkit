package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvkit/internal/cv"
)

var exportFormat string

// formatExt maps an export format to its renderer and file extension.
var formatExt = map[string]struct {
	ext    string
	render func(*cv.CV) string
}{
	"tex":      {".tex", (*cv.CV).RenderLaTeX},
	"md":       {".md", (*cv.CV).RenderMarkdown},
	"markdown": {".md", (*cv.CV).RenderMarkdown},
	"txt":      {".txt", (*cv.CV).RenderText},
	"text":     {".txt", (*cv.CV).RenderText},
}

var exportCmd = &cobra.Command{
	Use:   "export [input] [output]",
	Short: "Render a CV JSON to a chosen format (tex, md, txt)",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, ok := formatExt[strings.ToLower(exportFormat)]
		if !ok {
			return fmt.Errorf("unknown format %q (want: tex, md, txt)", exportFormat)
		}
		in := argOrDefault(args, 0, "cv_data.json")
		out := argOrDefault(args, 1, deriveName(in, f.ext))

		c, err := cv.Load(in)
		if err != nil {
			return err
		}
		if err := os.WriteFile(out, []byte(f.render(c)), 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s  ✓\n", out)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "md", "output format: tex, md, or txt")
	rootCmd.AddCommand(exportCmd)
}

// deriveName maps the input name to an output name with the given extension,
// applying the same cv_data[_role] -> cv[_role] convention as build.
func deriveName(in, ext string) string {
	base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
	switch {
	case base == "cv_data":
		base = "cv"
	case strings.HasPrefix(base, "cv_data_"):
		base = "cv_" + strings.TrimPrefix(base, "cv_data_")
	}
	return filepath.Join(filepath.Dir(in), base+ext)
}
