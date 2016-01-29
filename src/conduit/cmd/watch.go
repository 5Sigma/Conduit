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
	// "conduit/engine"
	"conduit/log"
	"conduit/queue"
	"github.com/spf13/cobra"
)

var s string

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Begin watching for commands.",
	Long: `Start processing the command queue. Conduit will run and wait for a
command to be delivered to it for processing.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting...")
		q := queue.GetQueue()
		for {
			script, err := q.Get()
			if err != nil {
				log.Error(err.Error())
			}
			if script != nil {
				log.Info(string(script.ScriptBody))
				// scriptBody, err := mgr.GetScript(cmd.RemoteScriptUrl)
				// if err != nil {
				// 	log.Error(err.Error())
				// } else {
				// 	err := engine.Execute(scriptBody)
				// 	if err != nil {
				// 		log.Error(err.Error())
				// 	} else {
				// 		q.Delete(cmd)
				// 	}
				// }
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// watchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// watchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
