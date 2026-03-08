package cmd

import (
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
		logger.Log.Info("starting")
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
		defer w.Close()
		fileEvents := make(chan string, 64)
		go w.Run(fileEvents)

		processor := events.NewProcessor(32)
		b := builder.New(build)
		r := runner.New(execCmd)

		go processor.Run(200 * time.Millisecond)
		go func() {
			for path := range fileEvents {
				processor.In <- events.Event{
					Path: path,
				}
			}
		}()
		for range processor.Out {
			logger.Log.Info("change detected")
			err := b.Build()
			if err != nil {
				logger.Log.Error("build failed", "error", err)
				continue
			}
			logger.Log.Info("restarting server")
			err = r.Restart()
			if err != nil {
				logger.Log.Error("restart failed", "error", err)
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
