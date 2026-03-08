package cmd

import (
	"fmt"
	"time"

	"github.com/parrothacker1/watchforge/internal/builder"
	"github.com/parrothacker1/watchforge/internal/events"
	"github.com/parrothacker1/watchforge/internal/runner"
	"github.com/parrothacker1/watchforge/internal/watcher"

	"github.com/spf13/cobra"
)

var root string
var build string
var execCmd string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run hot reload engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[watchforge] starting")
		w, err := watcher.New(root)
		if err != nil {
			return err
		}

		processor := events.NewProcessor(32)

		b := builder.New(build)
		r := runner.New(execCmd)

		go processor.Run(200 * time.Millisecond)
		w.Run()

		for range processor.Out {

			fmt.Println("[watchforge] change detected")

			err := b.Build()
			if err != nil {
				fmt.Println("[watchforge] build failed:", err)
				continue
			}

			fmt.Println("[watchforge] restarting server")

			err = r.Restart()
			if err != nil {
				fmt.Println("[watchforge] restart failed:", err)
			}
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&root, "root", ".", "root directory to watch")
	runCmd.Flags().StringVar(&build, "build", "", "build command")
	runCmd.Flags().StringVar(&execCmd, "exec", "", "execution command")

	rootCmd.AddCommand(runCmd)
}
