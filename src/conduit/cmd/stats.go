package cmd

import (
	"conduit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"postmaster/client"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Retrieve system statistics from the server",
	Long: `Gathers system statistics such as connected clients, and pending
message count from the remote server.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := client.Client{
			Host:  viper.GetString("host"),
			Token: viper.GetString("access_key"),
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
	RootCmd.AddCommand(statsCmd)
}
