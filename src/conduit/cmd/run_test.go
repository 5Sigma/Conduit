package cmd

import (
	"github.com/spf13/viper"
	"postmaster/mailbox"
	"testing"
)

func TestRun(t *testing.T) {
	mb, err := mailbox.Create("test.testrun")
	if err != nil {
		t.Fatal(err)
	}
	key := mailbox.AccessKey{MailboxId: mb.Id}
	err = key.Create()
	if err != nil {
		t.Fatal(err)
	}
	msg, err := mb.PutMessage(`$("test");`)
	if err != nil {
		t.Fatal(err)
	}
	viper.Set("host", ":5112")
	viper.Set("mailbox", mb.Id)
	viper.Set("access_key", key.Secret)
	viper.Set("access_key_name", key.Name)
	viper.Set("show_requests", true)
	go serverCmd.Run(serverCmd, []string{})
	runCmd.ParseFlags([]string{"-1", "-d"})
	runCmd.Run(runCmd, []string{})
	dep, _ := mailbox.FindDeployment(msg.Deployment)
	resp, _ := dep.GetResponses()
	if len(resp) != 1 {
		t.Fatal("No response")
	}
}
