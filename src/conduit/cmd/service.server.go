// +build windows

package cmd

import (
	"conduit/log"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows/svc"
	"postmaster/server"
	"time"
)

type conduitServerService struct{}

func (m *conduitServerService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	if viper.IsSet("enable_long_polling") {
		server.EnableLongPolling = viper.GetBool("enable_long_polling")
	}
	err := server.Start(viper.GetString("host"))
	fmt.Println("Could not start server:", err)
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			default:
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

// serviceCmd represents the service command
var serviceServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Conduit as a Windows service in Server mode.",
	Long: `Use this command line flag to run Conduit as a Windows service. The
service in Windows should be setup to run 'conduit service'.`,
	Run: func(cmd *cobra.Command, args []string) {
		run := svc.Run
		err := run("Conduit Server", &conduitService{})
		if err != nil {
			log.Fatal(err.Error())
		}

	},
}

func init() {
	serviceCmd.AddCommand(serviceServerCmd)
}
