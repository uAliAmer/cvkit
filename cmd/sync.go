package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvgen/internal/cv"
)

var syncCmd = &cobra.Command{
	Use:   "sync [input] [output]",
	Short: "Copy a validated CV JSON to the portfolio data path",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := argOrDefault(args, 0, "cv_data.json")
		out := argOrDefault(args, 1, filepath.Join("portfolio", "src", "lib", "cv.json"))
		raw, err := cv.ReadValidated(in)
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
	rootCmd.AddCommand(syncCmd)
}
