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
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register [name]",
	Short: "Register a new mailbox on the remote server.",
	Long: `This will register a new mailbox on the configured Conduit server. An
access key will also be generated and bound to the mailbox. This key can be
used by Conduit clients to receive messages from this mailbox.

Mailboxes can be locally generated as well.
Use "conduit help server register" for more information.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("No mailbox identifier specified.")
		}
		client, err := ClientFromConfig()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not configure client")
		}
		resp, err := client.RegisterMailbox(args[0])
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not register mailbox.")
		}
		log.Infof("Mailbox registered: %s", resp.Mailbox)
		log.Infof("Access key: %s", resp.AccessKeySecret)
	},
}

func init() {
	RootCmd.AddCommand(registerCmd)

}
