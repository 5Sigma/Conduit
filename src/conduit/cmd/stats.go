package cmd

import (
	"conduit/log"
	"conduit/queue"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Retrieve system statistics from the server",
	Long: `Gathers system statistics such as connected clients, and pending
message count from the remote server.`,
	Run: func(cmd *cobra.Command, args []string) {
		q := queue.New(viper.GetString("queue.host"), viper.GetString("mailbox"),
			viper.GetString("access_key"))
		stats, err := q.SystemStats()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not retrieve statistics")
		}
		log.Stats("Host", q.Client.Host)
		log.Stats("Pending messages", stats.PendingMessages)
		log.Stats("Connected clients", stats.ConnectedClients)
		log.Stats("Total mailboxes", stats.TotalMailboxes)
	},
}

func init() {
	RootCmd.AddCommand(statsCmd)
}
