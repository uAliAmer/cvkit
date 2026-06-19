package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvkit/internal/cv"
)

var syncForce bool

var syncCmd = &cobra.Command{
	Use:   "sync [input] [output]",
	Short: "Copy a validated CV JSON to the portfolio data path",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := argOrDefault(args, 0, "cv_data.json")
		out := argOrDefault(args, 1, filepath.Join("portfolio", "src", "lib", "cv.json"))

		// Parse + structurally validate before propagating to the site, so a
		// broken CV never reaches the portfolio. --force downgrades to a warning.
		c, err := cv.Load(in)
		if err != nil {
			return err
		}
		if problems := c.Validate(); len(problems) > 0 {
			for _, p := range problems {
				fmt.Fprintf(os.Stderr, "  ✗ %s\n", p)
			}
			if !syncForce {
				return fmt.Errorf("%s: %d problem(s); fix them or pass --force", in, len(problems))
			}
			fmt.Fprintln(os.Stderr, "  (--force: syncing anyway)")
		}

		// Copy the source bytes verbatim so formatting is preserved.
		raw, err := os.ReadFile(in)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(out, raw, 0o644); err != nil {
			return err
		}
		fmt.Printf("synced %s -> %s  ✓\n", in, out)
		return nil
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "sync even if validation finds problems")
	rootCmd.AddCommand(syncCmd)
}
