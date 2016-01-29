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
	"conduit/engine"
	"conduit/log"
	"conduit/queue"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math"
	"time"
)

var s string
var errorCount int

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Begin watching for commands.",
	Long: `Start processing the command queue. Conduit will run and wait for a
command to be delivered to it for processing.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting...")

		q := queue.New(viper.GetString("queue.host"), viper.GetString("mailbox"))

		for {
			script, err := q.Get()
			if err != nil {
				log.Warn(err.Error())
				log.Error("Could not poll for messages.")
				if errorCount < 15 {
					errorCount++
				}
				time.Sleep(time.Duration(math.Pow(float64(errorCount), 2)) * time.Second)
				continue
			}
			errorCount = 0
			if script != nil {
				err := engine.Execute(script.ScriptBody)
				if err != nil {
					log.Error("Error executing script." + script.Receipt)
					log.Debug(err.Error())
				}
				err = q.Delete(script)
				if err != nil {
					log.Error("Could not confirm script.")
					log.Debug(err.Error())
				} else {
					// log.Info("Script confirmed: " + script.Receipt)
				}
			} else {
				log.Info("No messages")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)
}
