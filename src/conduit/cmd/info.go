package cmd

import (
	"conduit/log"
	"fmt"
	"github.com/pivotal-golang/bytefmt"
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
		log.Alert("\nSystem Statistics\n")
		log.Stats("Host", client.Host)
		log.Stats("Version", stats.Version)
		if cmd.Flag("verbose").Value.String() == "true" {
			log.Stats("Total messages", stats.MessageCount)
		}
		log.Stats("Pending messages", stats.PendingMessages)
		log.Stats("Connected clients",
			fmt.Sprintf("%d / %d", stats.ConnectedClients, stats.TotalMailboxes))
		if cmd.Flag("verbose").Value.String() == "true" {
			log.Stats("Database version", stats.DBVersion)
			log.Stats("CPU's in use", stats.CPUCount)
			log.Stats("Active threads", stats.Threads)
			log.Stats("Memory in use", bytefmt.ByteSize(stats.MemoryAllocated))
			log.Stats("Lookups", fmt.Sprintf("%d", stats.Lookups))
			log.Stats("Next GC at", bytefmt.ByteSize(stats.NextGC))
			log.Stats("File store count", fmt.Sprintf("%d", stats.FileStoreCount))
			log.Stats("File store size", bytefmt.ByteSize(uint64(stats.FileStoreSize)))
		}
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolP("verbose", "v", false, "Show all details.")
}
