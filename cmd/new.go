package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uAliAmer/cvkit/internal/cv"
)

var (
	newFrom  string
	newForce bool
)

var roleSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

var newCmd = &cobra.Command{
	Use:   "new <role>",
	Short: "Scaffold a role-specific variant from a base CV JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := strings.Trim(roleSlugRe.ReplaceAllString(strings.ToLower(args[0]), "_"), "_")
		if slug == "" {
			return fmt.Errorf("role %q produced an empty slug", args[0])
		}

		// Validate the base before cloning so variants start from valid data.
		if _, err := cv.Load(newFrom); err != nil {
			return err
		}
		raw, err := os.ReadFile(newFrom)
		if err != nil {
			return err
		}

		out := filepath.Join(filepath.Dir(newFrom), "cv_data_"+slug+".json")
		if _, err := os.Stat(out); err == nil && !newForce {
			return fmt.Errorf("%s already exists; pass --force to overwrite", out)
		}
		if err := os.WriteFile(out, raw, 0o644); err != nil {
			return err
		}
		fmt.Printf("created %s  ✓\n", out)
		fmt.Printf("next: edit it for the %s role, then `cvkit build %s`\n", args[0], out)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&newFrom, "from", "cv_data.json", "base CV JSON to clone")
	newCmd.Flags().BoolVar(&newForce, "force", false, "overwrite if the variant already exists")
	rootCmd.AddCommand(newCmd)
}
