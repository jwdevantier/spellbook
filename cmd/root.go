package cmd

import (
	"fmt"
	"github.com/jwdevantier/spellbook/utils"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)
const configName = ".spellbook"

func init() {
	cobra.OnInitialize(initConfig)
}

var Config *utils.Config

func initConfig() {
	home, err := homedir.Dir()
	if err != nil {
		return
	}
	config, err := utils.ReadConfig([]string{home, "."})
	if err != nil {
		fmt.Println("failed to read configs")
		fmt.Println(err)
		// TODO: abort, failed to read config
	}
	Config = config
}

var rootCmd = &cobra.Command{
	Use: "spellbook",
	Short: "Easy access to your best shell commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Main command run....")
		_ = cmd.Help()

	},
}

// Execute executes the CLI interface
func Execute() error {
	return rootCmd.Execute()
}