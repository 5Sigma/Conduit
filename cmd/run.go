package cmd

import (
	"errors"
	"github.com/5sigma/conduit/engine"
	"github.com/5sigma/conduit/log"
	"github.com/elazarl/goproxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math"
	"math/rand"
	"net/http"
	"time"
)

var s string
var errorCount int

func runClient(noLoop bool) {
	viper.SetDefault("script_timeout", 300)
	log.LogFile = true
	if viper.GetBool("master.enabled") {
		log.Info("Launching master server")
		go runProxy()
	}
	log.Info("Waiting for messages...")

	client, _ := MailboxClientFromConfig()

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
		time.Sleep(time.Duration(rand.Intn(1000)+2000) * time.Millisecond)
		resp, err := client.Get()

		// If an error is returned by the client we will begin an exponential back
		// off in retrying. The backoff caps out at 15 retries.
		if err != nil {
			log.Error("Error getting messages: " + err.Error())
			if errorCount < 15 {
				errorCount++
			}
			expBackoff := int(math.Pow(float64(errorCount), 2))
			displacement := rand.Intn(errorCount + 1)
			sleepTime := expBackoff + displacement
			time.Sleep(time.Duration(sleepTime) * time.Second)
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
			errChan := make(chan string, 1)
			timeoutSeconds := viper.GetInt("script_timeout")
			go func() {
				err = eng.Execute(resp.Body)
				if err != nil {
					errChan <- err.Error()
				} else {
					errChan <- ""
				}
			}()
			select {
			case e := <-errChan:
				if e != "" {
					err = errors.New(e)
				}
			case <-time.After(time.Second * time.Duration(timeoutSeconds)):
				log.Warn("Timing out script")
				err = errors.New("Scirpt timeed out")
			}
			executionTime := time.Since(executionStartTime)
			log.Infof("Script executed in %s", executionTime)
			if err != nil {
				log.Error("Error executing script " + resp.Message)
				log.Debug(err.Error())
				client.Respond(resp.Message, err.Error(), true)
			}
			_, err = client.Delete(resp.Message)
			if err != nil {
				log.Debug(err.Error())
				log.Error("Could not confirm script.")
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
