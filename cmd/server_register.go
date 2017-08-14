package cmd

import (
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"github.com/spf13/cobra"
)

// registerCmd represents the register command
var serverRegisterCmd = &cobra.Command{
	Use:   "register [name]",
	Short: "Register a new mailbox",
	Long:  `This registers a new mailbox for the local server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("No mailbox name specified")
		}
		mailbox.OpenDB()
		mb, err := mailbox.Create(args[0])
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not create mailbox")
		}
		key := &mailbox.AccessKey{MailboxId: mb.Id}
		err = key.Create()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not create mailbox access token")
		}
		log.Infof("Mailbox created: %s", mb.Id)
		log.Infof("Access key created: %s", key.Secret)
	},
}

func init() {
	serverCmd.AddCommand(serverRegisterCmd)

}
