package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvkit/internal/cv"
)

var diffCmd = &cobra.Command{
	Use:   "diff <base> <other>",
	Short: "Show what differs between two CV variants",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := cv.Load(args[0])
		if err != nil {
			return err
		}
		b, err := cv.Load(args[1])
		if err != nil {
			return err
		}
		lines := cv.Diff(a, b)
		if len(lines) == 0 {
			fmt.Printf("%s and %s are equivalent\n", args[0], args[1])
			return nil
		}
		fmt.Printf("%s -> %s\n", args[0], args[1])
		for _, l := range lines {
			fmt.Println(l)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
