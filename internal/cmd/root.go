package cmd

import (
	"github.com/spf13/cobra"
)

var version string

var rootCmd = &cobra.Command{
	Use:   "watchforge",
	Short: "Watchforge is a hot reload engine for Go projects",
	Long:  "Watchforge automatically rebuilds and restarts Go servers when files change.",
}

func Execute(v string) {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	version = v
	cobra.CheckErr(rootCmd.Execute())
}
