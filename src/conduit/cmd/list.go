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
	"strconv"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List past deployments",
	Long: `Lists past deployments, by default it will list the last 10 deployments
made by your access key.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := client.Client{
			Host:  viper.GetString("queue.host"),
			Token: viper.GetString("access_key"),
		}
		limitToken := (cmd.Flag("all").Value.String() == "false")
		count, _ := strconv.ParseInt(cmd.Flag("count").Value.String(), 10, 64)
		resp, err := client.ListDeploys(cmd.Flag("name").Value.String(),
			limitToken, int(count))
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
			log.Infof("   Deployed at: %s",
				dep.CreatedAt.Format("01/02 03:04 PM"))
			log.Infof("   Deployed by: %s", dep.DeployedBy)
			log.Infof("   Processed scripts: %d/%d",
				dep.MessageCount-dep.PendingCount, dep.MessageCount)
			log.Infof("   Repsonses: %d/%d", dep.ResponseCount, dep.MessageCount)
		}
	},
}

func init() {
	deployCmd.AddCommand(listCmd)
	listCmd.Flags().IntP("count", "c", 10,
		"The maximum number of deployments to return")
	listCmd.Flags().BoolP("all", "a", false, "Return all deployments")
	listCmd.Flags().StringP("name", "n", ".*",
		"A search pattern to limit deployment names")

}
