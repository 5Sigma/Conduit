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

// accessCmd represents the access command
var accessCmd = &cobra.Command{
	Use:   "access [name]",
	Short: "Generate an administrative access key for a conduit server.",
	Long: `This generates and returns a administrative API access key for the
local Conduit server. This key gives full access to the Conduit API and should
be used for administrative purposes in a Conduit client or by an external system
that can manage the Conduit service.

For audit purposes the access key can be given a name. If no name is specified a
randomly generated identifier will be used.`,
	Run: func(cmd *cobra.Command, args []string) {
		mailbox.OpenDB()
		key := mailbox.AccessKey{FullAccess: true}
		if len(args) > 0 {
			key.Name = args[0]
		}
		err := key.Create()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not create access key.")
		}
		log.Info("Access key created: ")
		log.Info("  Access key name: " + key.Name)
		log.Info("  Access key: " + key.Secret)
	},
}

func init() {
	serverCmd.AddCommand(accessCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accessCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accessCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
