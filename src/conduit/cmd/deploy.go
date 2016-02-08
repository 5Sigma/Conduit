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
			Host:    viper.GetString("host"),
			Mailbox: viper.GetString("mailbox"),
			Token:   viper.GetString("access_key"),
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
			log.Fatal("Bad script syntax")
			log.Fatal(err.Error())
		}
		res, err := eng.ExecuteFunction("$local", string(data))
		if err != nil {
			log.Fatal("Local execution error")
			log.Fatal(err.Error())
		}
		if res != "" && res != "undefined" {
			log.Info("Local: " + res)
		}
		resp, err := client.Put(mailboxes, pattern, string(data),
			cmd.Flag("name").Value.String())
		if err != nil {
			log.Debug(err.Error())
			log.Error("Could not deploy script")
		}
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Infof("Script deployed to %d mailboxes (%d bytes)",
			len(resp.Mailboxes), len(data))
		bar := uiprogress.AddBar(len(resp.Mailboxes))
		bar.AppendCompleted()
		bar.PrependElapsed()
		uiprogress.Start()
		stats, err := client.PollDeployment(resp.Deployment,
			func(stats *api.DeploymentStats) bool {
				messagesProcessed := stats.MessageCount - stats.PendingCount
				bar.Set(int(messagesProcessed))
				if stats.PendingCount == 0 {
					return false
				} else {
					return true
				}
			})
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			log.Debug(err.Error())
			log.Error("Could not get deployment results")
			return
		}
		if len(stats.Responses) > 0 {
			log.Info("\nResponses:")
		}
		for _, r := range stats.Responses {
			log.Infof("%s: %s", r.Mailbox, r.Response)
		}
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("pattern", "p", "", "Wildcard search for mailboxes.")
	deployCmd.Flags().StringP("name", "n", "", "Deployment name")
}
