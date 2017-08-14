package cmd

import (
	"github.com/5sigma/conduit/log"
	"github.com/spf13/cobra"
)

// deregisterCmd represents the deregister command
var deregisterCmd = &cobra.Command{
	Use:   "deregister",
	Short: "Deregister a mailbox",
	Long:  `Remove a mailbox, its access tokens, and messages.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("No mailbox identifier specified.")
		}
		client, err := AdminClientFromConfig()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not configure client")
		}
		_, err = client.DeregisterMailbox(args[0])
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not remove mailbox.")
		}
		log.Infof("Mailbox deregistered: %s", args[0])
	},
}

func init() {
	RootCmd.AddCommand(deregisterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deregisterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deregisterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
