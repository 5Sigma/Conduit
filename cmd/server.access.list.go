package cmd

import (
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"github.com/spf13/cobra"
)

// acessListCmd represents the acessList command
var accessListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all administrative access keys",
	Long: `Display a list of all administrative access keys that have been
generated using 'conduit server access'. The server must be stopped when
running this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		mailbox.OpenDB()
		keys, err := mailbox.AdminKeys()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not list keys")
		}
		log.Alert("Access keys:")
		if len(keys) > 0 {
			for _, k := range keys {
				log.Info(k.Name)
			}
		} else {
			log.Warn("There are no access keys.")
		}
		mailbox.CloseDB()
	},
}

func init() {
	accessCmd.AddCommand(accessListCmd)
}
