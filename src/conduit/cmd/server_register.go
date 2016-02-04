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
		token, err := mailbox.CreateMailboxToken(mb.Id)
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not create mailbox access token")
		}
		log.Infof("Mailbox created: %s", mb.Id)
		log.Infof("Access key created: %s", token.Token)
	},
}

func init() {
	serverCmd.AddCommand(serverRegisterCmd)

}
