// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// purgeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// purgeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
