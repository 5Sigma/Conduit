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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverAccessRevokeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverAccessRevokeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
