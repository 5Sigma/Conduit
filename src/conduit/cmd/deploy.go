package cmd

import (
	"conduit/engine"
	"conduit/log"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"postmaster/api"
	"postmaster/client"
	"time"
)

var deployTimeout int

// sendCmd represents the send command
var deployCmd = &cobra.Command{
	Use:     "deploy [file] [client]",
	Aliases: []string{"send"},
	Short:   "Send a script to be executed",
	Long: `Send a script to be executed by a client. The file can either be a
javascript file or a zip file containing a javascript file and other arbitrary
files.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := client.Client{
			Host:          viper.GetString("host"),
			Mailbox:       viper.GetString("mailbox"),
			AccessKey:     viper.GetString("access_key"),
			AccessKeyName: viper.GetString("access_key_name"),
			ShowRequests:  viper.GetBool("show_requests"),
		}
		if len(args) == 0 {
			log.Fatal("No script specified.")
		}
		filename := args[0]
		mailboxes := args[1:]
		pattern := cmd.Flag("pattern").Value.String()
		if pattern == "" && len(mailboxes) == 0 {
			log.Fatal("Must provide either a list of mailboxes, a pattern, or both.")
		}
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err.Error())
		}
		eng := engine.New()
		err = eng.Validate(string(data))
		if err != nil {
			log.Error(err.Error())
			log.Fatal("Bad script syntax")
		}
		res, err := eng.ExecuteFunction("$local", string(data))
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Local execution error")
		}
		if res != "" && res != "undefined" {
			log.Info("Local: " + res)
		}
		resp, err := client.Put(mailboxes, pattern, string(data),
			cmd.Flag("name").Value.String(), cmd.Flag("attach").Value.String())
		if err != nil {
			log.Debug(err.Error())
			log.Error("Could not deploy script")
		}
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Infof("Script deployed to %d mailboxes (%d bytes)",
			len(resp.Mailboxes), len(data))
		log.Infof("Deployment name: %s", resp.Deployment)

		if cmd.Flag("no-results").Value.String() == "true" {
			return
		}

		bar := uiprogress.AddBar(len(resp.Mailboxes))
		bar.AppendCompleted()
		bar.PrependElapsed()
		uiprogress.Start()
		var pollStart = time.Now()
		stats, err := client.PollDeployment(resp.Deployment,
			func(stats *api.DeploymentStats) bool {
				messagesProcessed := stats.MessageCount - stats.PendingCount
				bar.Set(int(messagesProcessed))
				if time.Since(pollStart) > time.Duration(deployTimeout)*time.Second {
					return false
				}
				if stats.PendingCount == 0 {
					return false
				} else {
					return true
				}
			})
		uiprogress.Stop()
		if err != nil {
			log.Debug(err.Error())
			log.Error("Could not get deployment results")
			return
		}
		if len(stats.Responses) > 0 {
			log.Alert("\nResponses:")
		}
		stats.Responses.Sort()
		displayResponses(stats.Responses,
			cmd.Flag("expand").Value.String() != "true")
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("pattern", "p", "",
		"Wildcard search for mailboxes.")
	deployCmd.Flags().StringP("name", "n", "",
		"A custom name for this deployment.")
	deployCmd.Flags().BoolP("no-results", "x", false,
		"Dont poll for responses.")
	deployCmd.Flags().StringP("attach", "a", "",
		"Attach a file asset to this deployment.")
	deployCmd.Flags().IntVarP(&deployTimeout, "timeout", "t", 20,
		"Response polling timeout.")
	deployCmd.Flags().BoolP("expand", "e", false,
		"Expand results (don't consolidate)")
}
