package cmd

import (
	"fmt"
	"github.com/5sigma/conduit/engine"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var executeCmd = &cobra.Command{
	Use:     "execute [script]",
	Short:   "Execute a script file.",
	Long:    `Execute a Javascript file at the specified path.`,
	Aliases: []string{"exec"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No file specified.")
			os.Exit(-1)
		}
		startTime := time.Now()
		eng := engine.New()
		eng.ExecuteFile(args[0])
		fmt.Printf("Script executed in %s .\n", time.Since(startTime))
	},
}

func init() {
	RootCmd.AddCommand(executeCmd)
}
