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
		for _, k := range keys {
			log.Info(k.Name)
		}
		mailbox.CloseDB()
	},
}

func init() {
	accessCmd.AddCommand(accessListCmd)
}
