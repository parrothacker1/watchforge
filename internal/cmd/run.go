package cmd

import (
	"fmt"
	"time"

	"github.com/parrothacker1/watchforge/internal/builder"
	"github.com/parrothacker1/watchforge/internal/events"
	"github.com/parrothacker1/watchforge/internal/files"
	"github.com/parrothacker1/watchforge/internal/logger"
	"github.com/parrothacker1/watchforge/internal/runner"
	"github.com/parrothacker1/watchforge/internal/watcher"

	"github.com/spf13/cobra"
)

var root string
var build string
var execCmd string
var useGitignore bool
var debug bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run hot reload engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Init(debug)
		logger.Log.Info("[watchforge] starting")
		var filter *files.Filter
		if useGitignore {
			var err error
			filter, err = files.New(root, nil)
			if err != nil {
				return err
			}
		}
		w, err := watcher.New(root, filter)
		if err != nil {
			return err
		}
		var fileEvents chan string
		w.Run(fileEvents)
		defer w.Close()

		processor := events.NewProcessor(32)

		b := builder.New(build)
		r := runner.New(execCmd)

		go processor.Run(200 * time.Millisecond)

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
	runCmd.Flags().BoolVar(&useGitignore, "gitignore", true, "respect .gitignore")
	runCmd.Flags().BoolVar(&debug, "debug", false, "enable debug logging")

	rootCmd.AddCommand(runCmd)
}
