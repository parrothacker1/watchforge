package cmd

import (
	"os"

	"github.com/parrothacker1/watchforge/internal/config"
	"github.com/parrothacker1/watchforge/internal/logger"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate watchforge.toml config",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Init(false)
		path := "watchforge.toml"
		if _, err := os.Stat(path); err == nil {
			logger.Log.Warn("config already exists", "path", path)
			return nil
		}
		cfg := config.Default()
		err := config.Write(path, cfg)
		if err != nil {
			return err
		}
		logger.Log.Info("config generated", "path", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
