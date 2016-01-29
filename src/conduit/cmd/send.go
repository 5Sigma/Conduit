package cmd

import (
	"conduit/log"
	"conduit/queue"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send [file] [client]",
	Short: "Send a script to be executed",
	Long: `Send a script to be executed by a client. The file can either be a
javascript file or a zip file containing a javascript file and other arbitrary
files.`,
	Run: func(cmd *cobra.Command, args []string) {
		q := queue.New(viper.GetString("queue.host"), viper.GetString("mailbox"))
		filename := args[0]
		client := args[1]
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = q.Put([]string{client}, &queue.ScriptCommand{ScriptBody: string(data)})
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Infof("Script deployed to %s (%d)", client, len(data))
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
