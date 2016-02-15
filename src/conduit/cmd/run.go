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
	"github.com/elazarl/goproxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math"
	"net/http"
	"time"
)

var s string
var errorCount int

func runClient(noLoop bool) {
	log.LogFile = true
	if viper.GetBool("master.enabled") {
		log.Info("Launching master server")
		go runProxy()
	}
	log.Info("Waiting for messages...")

	client, _ := ClientFromConfig()

	if viper.IsSet("master.host") {
		client.UseProxy = true
		client.ProxyAddress = "http://" + viper.GetString("master.host")
	}

	var persistantScripts = []*engine.ScriptEngine{}

	if viper.IsSet("agents") {
		engine.Agents = viper.GetStringMapString("agents")
		engine.AgentAccessKey = viper.GetString("access_key")
	}

	// Begin polling cycle
	for {
		resp, err := client.Get()

		// If an error is returned by the client we will begin an exponential back
		// off in retrying. The backoff caps out at 15 retries.
		if err != nil {
			log.Error(err.Error())
			if errorCount < 15 {
				errorCount++
			}
			time.Sleep(time.Duration(math.Pow(float64(errorCount), 2)) * time.Second)
			continue
		}

		// A response was received but it might be an empty response from the
		// server timing out the long poll.
		errorCount = 0
		if resp.Body != "" {
			log.Infof("Script receieved (%s)", resp.Message)
			eng := engine.New()
			eng.Constant("DEPLOYMENT_ID", resp.Deployment)
			eng.Constant("SCRIPT_ID", resp.Message)

			persistant, _ := eng.GetVar("$persistant", resp.Body)
			if p, ok := persistant.(bool); ok {
				if p {
					persistantScripts = append(persistantScripts, eng)
				}
			}

			if resp.Asset != "" {
				assetPath, err := client.DownloadAsset(resp.Asset)
				if err != nil {
					client.Respond(resp.Message, "Could not download asset", true)
					log.Error("Could not download asset")
					_, err = client.Delete(resp.Message)
					continue
				} else {
					log.Infof("Downloaded asset to %s", assetPath)
				}
				eng.SetAsset(assetPath)
			}

			executionStartTime := time.Now()
			err := eng.Execute(resp.Body)
			executionTime := time.Since(executionStartTime)
			log.Infof("Script executed in %s", executionTime)
			if err != nil {
				log.Error("Error executing script " + resp.Message)
				log.Debug(err.Error())
				client.Respond(resp.Message, err.Error(), true)
			}
			_, err = client.Delete(resp.Message)
			if err != nil {
				log.Error("Could not confirm script.")
				log.Debug(err.Error())
			} else {
				log.Debug("Script confirmed: " + resp.Message)
			}
		}
		if noLoop == true {
			break
		}
	}
}

func runProxy() {
	proxy := goproxy.NewProxyHttpServer()
	proxyAddress := viper.GetString("master.Address")
	err := http.ListenAndServe(proxyAddress, proxy)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// runCmd starts a Conduit client in polling mode. It will poll the server for
// messages and evaluate message bodies as scripts.
var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"client"},
	Short:   "Run Conduit in client mode",
	Long: `Start processing the command queue. Conduit will run and wait for a
command to be delivered to it for processing.`,
	Run: func(cmd *cobra.Command, args []string) {
		breakLoop := (cmd.Flag("one").Value.String() == "true")
		runClient(breakLoop)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("one", "1", false, "Process a single message and exit.")
}
