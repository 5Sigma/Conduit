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
	"github.com/spf13/viper"
	"postmaster/client"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := client.Client{
			Host:  viper.GetString("queue.host"),
			Token: viper.GetString("access_key"),
		}
		resp, err := client.ListDeploys()
		if err != nil {
			log.Debug(err.Error())
			log.Error("Could not list deploys")
			return
		}
		if len(resp.Deployments) == 0 {
			log.Warn("There are no open deployments")
		}
		for _, dep := range resp.Deployments {
			log.Infof("%s:", dep.Name)
			log.Infof("   Pending scripts: %d/%d", dep.PendingCount, dep.MessageCount)
			log.Infof("   Repsonses: %d", dep.ResponseCount)
		}
	},
}

func init() {
	deployCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
