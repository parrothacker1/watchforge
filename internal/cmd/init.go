package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a watchforge configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {

		config := `root = "."

build = "go build -o ./bin/server ./cmd/server"

exec = "./bin/server"

debounce = 200
`

		if _, err := os.Stat(".watchforge.toml"); err == nil {
			return fmt.Errorf(".watchforge.toml already exists")
		}

		err := os.WriteFile(".watchforge.toml", []byte(config), 0644)
		if err != nil {
			return err
		}

		fmt.Println("Created .watchforge.toml")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
