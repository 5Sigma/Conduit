// +build windows

package cmd

import (
	"github.com/5sigma/conduit/log"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"time"
)

type conduitService struct{}

func (m *conduitService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	go runClient(false)
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
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run Conduit as a Windows service.",
	Long: `Use this command line flag to run Conduit as a Windows service. The
service in Windows should be setup to run 'conduit service'.`,
	Run: func(cmd *cobra.Command, args []string) {
		run := svc.Run
		err := run("Conduit Client", &conduitService{})
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(serviceCmd)
}
