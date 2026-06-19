package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvkit/internal/cv"
)

var (
	tailorJD  string
	tailorTop int
)

var tailorCmd = &cobra.Command{
	Use:   "tailor [input]",
	Short: "Match a CV against a job description and suggest what to surface",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if tailorJD == "" {
			return fmt.Errorf("--jd <file> is required")
		}
		jd, err := os.ReadFile(tailorJD)
		if err != nil {
			return err
		}
		in := argOrDefault(args, 0, "cv_data.json")
		c, err := cv.Load(in)
		if err != nil {
			return err
		}
		r := c.Tailor(string(jd), tailorTop)

		fmt.Println("Matched keywords (already in your CV — lead with these):")
		printTerms(r.Matched)
		fmt.Println("\nGaps (in the JD, missing from your CV — add if true):")
		printTerms(r.Gaps)
		fmt.Println("\nEntries by relevance (surface the top ones):")
		for _, e := range r.Surface {
			fmt.Printf("  %2d  %s\n", e.Hits, e.Name)
		}
		return nil
	},
}

func printTerms(t []cv.TermCount) {
	if len(t) == 0 {
		fmt.Println("  (none)")
		return
	}
	for _, tc := range t {
		fmt.Printf("  %2d  %s\n", tc.Count, tc.Term)
	}
}

func init() {
	tailorCmd.Flags().StringVar(&tailorJD, "jd", "", "path to a job-description text file (required)")
	tailorCmd.Flags().IntVar(&tailorTop, "top", 15, "max keywords to show per list")
	rootCmd.AddCommand(tailorCmd)
}
