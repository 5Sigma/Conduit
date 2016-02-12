package cmd

import (
	"conduit/log"
	"github.com/spf13/cobra"
	"strconv"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List past deployments",
	Long: `Lists past deployments, by default it will list the last 10 deployments
made by your access key.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ClientFromConfig()
		if err != nil {
			log.Debug(err.Error())
			log.Fatal("Could not configure client")
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
			log.Alertf("%s:", dep.Name)
			log.Infof("   Deployed at: %s",
				dep.CreatedAt.Format("01/02 03:04 PM"))
			log.Infof("   Deployed by: %s", dep.DeployedBy)
			log.Infof("   Executions: %d/%d",
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
