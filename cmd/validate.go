package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvkit/internal/cv"
)

var checkLinks bool

var validateCmd = &cobra.Command{
	Use:   "validate [input]",
	Short: "Check a CV JSON for missing fields and malformed entries",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := argOrDefault(args, 0, "cv_data.json")
		c, err := cv.Load(in)
		if err != nil {
			return err
		}
		problems := c.Validate()
		if checkLinks {
			problems = append(problems, c.CheckLinks()...)
		}
		if len(problems) == 0 {
			fmt.Printf("%s is valid  ✓\n", in)
			return nil
		}
		// Non-nil error => non-zero exit (V8).
		for _, p := range problems {
			fmt.Printf("  ✗ %s\n", p)
		}
		return fmt.Errorf("%s: %d problem(s)", in, len(problems))
	},
}

func init() {
	validateCmd.Flags().BoolVar(&checkLinks, "links", false, "also HTTP-check every project link")
	rootCmd.AddCommand(validateCmd)
}
