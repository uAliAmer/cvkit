package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var keepTex bool

var pdfCmd = &cobra.Command{
	Use:   "pdf [input]",
	Short: "Build a CV JSON to LaTeX and compile it to PDF with XeLaTeX",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bin, err := exec.LookPath("xelatex")
		if err != nil {
			return fmt.Errorf("xelatex not found in PATH; install a TeX distribution (TeX Live / MacTeX) to use 'pdf'")
		}

		in := argOrDefault(args, 0, "cv_data.json")
		tex := deriveTexName(in)
		if err := buildOne(in, tex); err != nil {
			return err
		}

		dir := filepath.Dir(tex)
		// XeLaTeX writes outputs to -output-directory; run from there so
		// relative asset lookups behave.
		c := exec.Command(bin, "-interaction=nonstopmode", "-halt-on-error",
			"-output-directory", dir, filepath.Base(tex))
		c.Dir = dir
		out, runErr := c.CombinedOutput()
		if runErr != nil {
			fmt.Fprintln(os.Stderr, string(out))
			return fmt.Errorf("xelatex failed: %w", runErr)
		}

		base := strings.TrimSuffix(filepath.Base(tex), ".tex")
		pdf := filepath.Join(dir, base+".pdf")
		// Clean aux artifacts; optionally the .tex too.
		for _, ext := range []string{".aux", ".log", ".out"} {
			os.Remove(filepath.Join(dir, base+ext))
		}
		if !keepTex {
			os.Remove(tex)
		}
		fmt.Printf("wrote %s  ✓\n", pdf)
		return nil
	},
}

func init() {
	pdfCmd.Flags().BoolVar(&keepTex, "keep-tex", false, "keep the intermediate .tex file")
	rootCmd.AddCommand(pdfCmd)
}
