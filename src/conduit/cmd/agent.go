package cmd

import (
	"conduit/agent"
	"conduit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run Conduit in agent mode",
	Long: `Agent mode allows one client instance to dispatch commands to other
instances. These other instances do not have their own mailbox. Instead commands
are sent to them from inside scripts executed by the handler instance. Scripts
can request agents to perform actions using the $agent function. For example:

$agent("agentName", function() {
	$shell("restart.sh");
});

Agents must be specified in their handler's configuration file. They are defined
using name/address map as follows:

agents:
	myAgentName: 10.0.100.15:2222

To run as an agent an address and port must be specified in the config file.
This is the interface and port to listen on. It should match the configuration
on the handler. For example:

agent_host: 10.0.100.15:2222

Agents must also have the same access_key as their handler in the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		agent.Address = viper.GetString("agent_host")
		agent.AccessKey = viper.GetString("access_key")
		if agent.Address == "" {
			cmd.Help()
			log.Fatal("\nThe 'agent_host' field in the config file is not present.")
		}
		log.Info("Starting agent on " + agent.Address)
		agent.Start()
	},
}

func init() {
	RootCmd.AddCommand(agentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// agentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// agentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
