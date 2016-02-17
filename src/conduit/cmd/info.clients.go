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
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// clientsCmd represents the clients command
var infoClientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Get the status of all mailboxes.",
	Long:  `Returns the connection status of all mailboxes.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ClientFromConfig()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not configure client")
		}
		stats, err := client.ClientStatus()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not retrieve statistics")
		}
		for _, st := range stats {
			versionStr := ""
			if cmd.Flag("info").Value.String() == "true" {
				if st.Version == "" {
					versionStr = "[ ? ]"
				} else {
					versionStr = fmt.Sprintf("[%s]", st.Version)
				}
			}
			kStr := fmt.Sprintf("%s %s", versionStr, st.Mailbox)
			if cmd.Flag("info").Value.String() == "true" {
				kStr += fmt.Sprintf(" (%s)", st.Host)
			}
			var vStr = ""
			if st.Online {
				vStr = "ONLINE "
			} else {
				vStr = "OFFLINE"
				if st.LastSeen.IsZero() {
					kStr += "  - Never checked in"
				} else {
					kStr += "  - last seen " + humanize.Time(st.LastSeen)
				}
			}
			if cmd.Flag("offline").Value.String() == "false" || !st.Online {
				log.Status(kStr, vStr, st.Online == true)
			}
		}
	},
}

func init() {
	infoCmd.AddCommand(infoClientsCmd)

	infoClientsCmd.Flags().BoolP("offline", "x", false, "Show only offline clients")
	infoClientsCmd.Flags().BoolP("info", "i", false, "Show addtional client information")
}
