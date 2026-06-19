package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [input]",
	Short: "Rebuild the .tex whenever the CV JSON changes",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := argOrDefault(args, 0, "cv_data.json")
		out := deriveTexName(in)

		w, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer w.Close()

		// Watch the directory, not the file: editors often replace files
		// (rename/create) rather than writing in place, which drops a direct
		// file watch.
		dir := filepath.Dir(in)
		if dir == "" {
			dir = "."
		}
		if err := w.Add(dir); err != nil {
			return err
		}

		if err := buildOne(in, out); err != nil {
			fmt.Fprintln(os.Stderr, "build error:", err)
		}
		fmt.Printf("watching %s ... (Ctrl-C to stop)\n", in)

		target := filepath.Clean(in)
		var last time.Time
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return nil
				}
				if filepath.Clean(ev.Name) != target {
					continue
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				// Debounce: editors emit several events per save.
				if time.Since(last) < 200*time.Millisecond {
					continue
				}
				last = time.Now()
				if err := buildOne(in, out); err != nil {
					fmt.Fprintln(os.Stderr, "build error:", err)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return nil
				}
				fmt.Fprintln(os.Stderr, "watch error:", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
