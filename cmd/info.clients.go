package cmd

import (
	"fmt"
	"github.com/5sigma/conduit/log"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// clientsCmd represents the clients command
var infoClientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Get the status of all mailboxes.",
	Long:  `Returns the connection status of all mailboxes.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := AdminClientFromConfig()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not configure client")
		}

		stats, err := client.ClientStatus()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not retrieve statistics")
		}
		stats.Sort()

		for _, st := range stats {
			versionStr := ""
			if cmd.Flag("verbose").Value.String() == "true" {
				if st.Version == "" {
					versionStr = "[ ? ]"
				} else {
					versionStr = fmt.Sprintf("[%s]", st.Version)
				}
			}
			kStr := fmt.Sprintf("%s %s", versionStr, st.Mailbox)
			if cmd.Flag("verbose").Value.String() == "true" {
				kStr += fmt.Sprintf(" (%s)", st.Host)
			}
			var vStr = ""
			if st.Online {
				vStr = "ONLINE "
			} else {
				vStr = "OFFLINE"
				if st.LastSeen.IsZero() {
					kStr += "  - Never checked in"
				} else {
					kStr += "  - last seen " + humanize.Time(st.LastSeen)
				}
			}
			if cmd.Flag("offline").Value.String() == "false" || !st.Online {
				log.Status(kStr, vStr, st.Online == true)
			}
		}
	},
}

func init() {
	infoCmd.AddCommand(infoClientsCmd)

	infoClientsCmd.Flags().BoolP("offline", "x", false,
		"Show only offline clients")
	infoClientsCmd.Flags().BoolP("verbose", "v", false,
		"Show addtional client information")
}
