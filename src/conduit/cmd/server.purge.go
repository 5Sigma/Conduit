package cmd

import (
	"conduit/log"
	"github.com/spf13/cobra"
	"postmaster/mailbox"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge [mailbox]",
	Short: "Purge all messages for a mailbox.",
	Long: `Delete all messages for the local server for a given mailbox. This
purges local data from the server's database. To purge a remote server use
conduit purge instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("No mailbox specified.")
		}
		mailbox.OpenDB()
		mailboxId := args[0]
		mb, err := mailbox.Find(mailboxId)
		if err != nil {
			log.Fatal("Could not lookup mailbox.")
			log.Debug(err.Error())
		}
		if mb == nil {
			log.Fatal("Could not find the mailbox specified")
		}
		c, err := mb.Purge()
		if err != nil {
			log.Fatal("Could not purge mailbox")
			log.Debug(err.Error())
		}
		log.Infof("Mailbox purged of %d messages.", c)
	},
}

func init() {
	serverCmd.AddCommand(purgeCmd)
}
