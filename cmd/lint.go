package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvkit/internal/cv"
)

var lintStrict bool

var lintCmd = &cobra.Command{
	Use:   "lint [input]",
	Short: "Flag content-quality issues: weak verbs, missing metrics, passive voice",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := argOrDefault(args, 0, "cv_data.json")
		c, err := cv.Load(in)
		if err != nil {
			return err
		}
		findings := c.Lint()
		if len(findings) == 0 {
			fmt.Printf("%s: no lint findings  ✓\n", in)
			return nil
		}
		for _, f := range findings {
			fmt.Printf("  • %s\n", f)
		}
		fmt.Printf("%d suggestion(s)\n", len(findings))
		if lintStrict {
			return fmt.Errorf("%d lint finding(s)", len(findings))
		}
		return nil
	},
}

func init() {
	lintCmd.Flags().BoolVar(&lintStrict, "strict", false, "exit non-zero if there are any findings")
	rootCmd.AddCommand(lintCmd)
}
