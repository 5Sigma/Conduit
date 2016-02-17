package cmd

import (
	"conduit/log"
	"github.com/spf13/cobra"
	"postmaster/api"
)

func getDeploymentInfo(depId string) {
	client, _ := ClientFromConfig()
	stats, err := client.PollDeployment(depId,
		func(stats *api.DeploymentStats) bool {
			return false
		})
	if err != nil {
		log.Debug(err.Error())
		log.Error("Could not get deployment results")
		return
	}
	log.Info("")
	log.Alert(depId)
	log.Info("")
	log.Infof("Total messages: %d", stats.MessageCount)
	log.Infof("Pending messages: %d", stats.PendingCount)
	log.Infof("Total responses: %d", stats.ResponseCount)
	if len(stats.Responses) > 0 {
		log.Alert("\nResponses:")
	}
	for _, r := range stats.Responses {
		if r.IsError == true {
			log.Errorf("%s: %s", r.Mailbox, r.Response)
		} else {
			log.Infof("%s: %s", r.Mailbox, r.Response)
		}
	}
	log.Info("")
}

// deployGetCmd represents the deployGet command
var deployGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get deployment information",
	Long: `Get deployment responses and statistics. You must specify the
deployment name. This name can be set when the deployment is made using the
"--name" flag. If a name was not specified a random name was assigned to it.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		} else {
			getDeploymentInfo(args[0])
		}
	},
}

func init() {
	deployCmd.AddCommand(deployGetCmd)
}
