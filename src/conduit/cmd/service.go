// +build windows

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run Conduit as a Windows service.",
	Long: `Use this command line flag to run Conduit as a Windows service. The
service in Windows should be setup to run 'conduit service'.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("service called")
	},
}

func init() {
	RootCmd.AddCommand(serviceCmd)
}
