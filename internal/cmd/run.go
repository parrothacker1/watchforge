package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var root string
var build string
var execCmd string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run hot reload engine",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("running watchforge", version)
		fmt.Println("root:", root)
		fmt.Println("build:", build)
		fmt.Println("exec:", execCmd)
	},
}

func init() {
	runCmd.Flags().StringVar(&root, "root", ".", "root directory to watch")
	runCmd.Flags().StringVar(&build, "build", "", "build command")
	runCmd.Flags().StringVar(&execCmd, "exec", "", "execution command")

	rootCmd.AddCommand(runCmd)
}
