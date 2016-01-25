package cmd

import (
	"conduit/queue"
	"conduit/storage"
	"fmt"
	"github.com/spf13/cobra"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send [file] [client]",
	Short: "Send a script to be executed",
	Long: `Send a script to be executed by a client. The file can either be a
javascript file or a zip file containing a javascript file and other arbitrary
files.`,
	Run: func(cmd *cobra.Command, args []string) {
		q := queue.GetQueue()
		filename := args[0]
		client := args[1]
		storage := storage.GetStorage()
		err := storage.PutScript(filename)
		if err != nil {
			fmt.Println(err)
		}
		script := &queue.ScriptCommand{
			RemoteScriptUrl: filename,
			RemoteAssets:    []string{},
		}
		err = q.Put(client, script)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(sendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
