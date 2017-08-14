package cmd

import (
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"github.com/spf13/cobra"
)

// serverAccessRevokeCmd represents the serverAccessRevoke command
var serverAccessRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke an access key",
	Long: `This will remove an access key from the database. It will no longer
be able to be used. The server must be stopped to perform this operation.`,
	Example: "conduit server access revoke mykey",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		mailbox.OpenDB()
		err := mailbox.Revoke(args[0])
		if err != nil {
			log.Fatal(err.Error())
		} else {
			log.Info("Access key revoked")
		}
		mailbox.CloseDB()
	},
}

func init() {
	accessCmd.AddCommand(serverAccessRevokeCmd)
}
