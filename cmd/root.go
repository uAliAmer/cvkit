// Package cmd wires up the cvgen CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cvgen",
	Short: "Generate a CV from one JSON source into multiple formats",
	Long: "cvgen turns a single cv_data.json into a LaTeX resume, portfolio JSON,\n" +
		"and more. One source of truth, many outputs.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
