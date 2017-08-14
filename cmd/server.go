package cmd

import (
	"fmt"
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Conduit in server mode.",
	Long: `Run the conduit message server. To manage the server use the server sub
commands. For help run 'conduit help server'.`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.IsSet("enable_long_polling") {
			server.EnableLongPolling = viper.GetBool("enable_long_polling")
		}
		log.LogFile = true
		err := server.Start(viper.GetString("host"))
		fmt.Println("Could not start server:", err)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
