package cmd

import (
	"conduit/log"
	"conduit/queue"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"postmaster/api"
	"time"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send [file] [client]",
	Short: "Send a script to be executed",
	Long: `Send a script to be executed by a client. The file can either be a
javascript file or a zip file containing a javascript file and other arbitrary
files.`,
	Run: func(cmd *cobra.Command, args []string) {
		q := queue.New(viper.GetString("queue.host"), viper.GetString("mailbox"),
			viper.GetString("access_key"))
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
		scriptCmd := &queue.ScriptCommand{ScriptBody: string(data)}
		resp, err := q.Client.Put(mailboxes, pattern, scriptCmd.ScriptBody)
		if err != nil {
			log.Debug(err.Error())
			log.Error("Could not deploy script")
		}
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Infof("Script deployed to %d mailboxes (%d bytes)",
			len(resp.Mailboxes), len(scriptCmd.ScriptBody))
		bar := uiprogress.AddBar(len(resp.Mailboxes))
		bar.AppendCompleted()
		bar.PrependElapsed()
		uiprogress.Start()
		stats, err := q.Client.PollDeployment(resp.Deployment,
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
			log.Error("Could not get deployment statistics.")
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
	RootCmd.AddCommand(sendCmd)
	sendCmd.Flags().StringP("pattern", "p", "", "Wildcard search for mailboxes.")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
