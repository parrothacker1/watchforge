package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/parrothacker1/watchforge/internal/builder"
	"github.com/parrothacker1/watchforge/internal/config"
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

func buildActionMaps() (map[string]struct{}, map[string]struct{}) {
	cfg := config.GetConfig()
	buildMap := make(map[string]struct{})
	restartMap := make(map[string]struct{})
	for _, ext := range cfg.Actions.Build {
		buildMap[ext] = struct{}{}
	}
	for _, ext := range cfg.Actions.Restart {
		restartMap[ext] = struct{}{}
	}
	return buildMap, restartMap
}

func classifyEvents(
	paths []string,
	buildMap map[string]struct{},
	restartMap map[string]struct{},
) (bool, bool) {
	var buildNeeded bool
	var restartNeeded bool
	for _, path := range paths {
		ext := filepath.Ext(path)
		if _, ok := buildMap[ext]; ok {
			buildNeeded = true
		}
		if _, ok := restartMap[ext]; ok {
			restartNeeded = true
		}
	}
	return buildNeeded, restartNeeded
}

func configChanged(paths []string) bool {
	for _, p := range paths {
		if filepath.Base(p) == "watchforge.toml" {
			return true
		}
	}
	return false
}

func runScheduler(
	processor *events.Processor,
	fileEvents chan string,
	b *builder.Builder,
	r *runner.Runner,
	buildMap map[string]struct{},
	restartMap map[string]struct{},
	ctx context.Context,
) {
	go func() {
		for path := range fileEvents {
			processor.In <- events.Event{Path: path}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("scheduler shutting down")
			return
		case batch, ok := <-processor.Out:
			if !ok {
				return
			}
			if configChanged(batch.Paths) {
				logger.Log.Info("config changed, reloading")
				cfg, err := config.Load(".")
				if err != nil {
					logger.Log.Error("failed to reload config", "error", err)
					continue
				}
				config.SetConfig(cfg)
				buildMap, restartMap = buildActionMaps()
				b = builder.New(ctx)
				r = runner.New(ctx)
				continue
			}
			buildNeeded, restartNeeded := classifyEvents(batch.Paths, buildMap, restartMap)
			if buildNeeded {
				b.Cancel()
				if err := b.Build(); err != nil {
					logger.Log.Error("build failed", "error", err)
					continue
				}
				r.Restart()
				continue
			}
			if restartNeeded {
				r.Restart()
			}
		}
	}
}

func setupEventProcessor() *events.Processor {
	cfg := config.GetConfig()
	processor := events.NewProcessor(32)
	go processor.Run(time.Duration(cfg.Watcher.Debounce) * time.Millisecond)
	return processor
}

func setupWatcher(root string, filter *files.Filter) (*watcher.Watcher, chan string, error) {
	w, err := watcher.New(root, filter)
	if err != nil {
		return nil, nil, err
	}
	fileEvents := make(chan string, 64)
	go w.Run(fileEvents)
	return w, fileEvents, nil
}

func createFilter() (*files.Filter, error) {
	cfg := config.GetConfig()
	if !cfg.Watcher.IgnoreGit {
		return nil, nil
	}
	return files.New(cfg.Root, cfg.Watcher.Ignore)
}

func loadConfig() *config.Config {
	cfg, err := config.Load(".")
	if err != nil {
		cfg = config.Default()
	}
	if root != "" {
		cfg.Root = root
	}
	if build != "" {
		cfg.Build.Command = build
	}
	if execCmd != "" {
		cfg.Run.Command = execCmd
	}
	if useGitignore {
		cfg.Watcher.IgnoreGit = useGitignore
	}
	if debug {
		cfg.Log.Debug = debug
	}
	return cfg
}

func run() error {
	cfg := loadConfig()
	config.SetConfig(cfg)
	logger.Init(cfg.Log.Debug)
	logger.Log.Info("starting")
	buildMap, restartMap := buildActionMaps()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		logger.Log.Info("shutdown signal received")
		cancel()
	}()

	filter, err := createFilter()
	if err != nil {
		return err
	}
	w, fileEvents, err := setupWatcher(cfg.Root, filter)
	if err != nil {
		return err
	}
	defer w.Close()

	processor := setupEventProcessor()
	b := builder.New(ctx)
	r := runner.New(ctx)
	defer r.Stop()
	runScheduler(processor, fileEvents, b, r, buildMap, restartMap, ctx)
	return nil
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run hot reload engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
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
