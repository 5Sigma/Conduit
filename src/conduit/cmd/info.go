package cmd

import (
	"conduit/log"
	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Retrieve system statistics from the server",
	Aliases: []string{"stats"},
	Long: `Gathers system statistics such as connected clients, and pending
message count from the remote server.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ClientFromConfig()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not configure client")
		}
		stats, err := client.Stats()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not retrieve statistics")
		}
		log.Stats("Host", client.Host)
		log.Stats("Pending messages", stats.PendingMessages)
		log.Stats("Connected clients", stats.ConnectedClients)
		log.Stats("Total mailboxes", stats.TotalMailboxes)
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
}
