package cmd

import (
	"github.com/spf13/viper"
	"os"
	"postmaster/mailbox"
	"testing"
)

func TestDeploy(t *testing.T) {
	mailbox.OpenMemDB()
	mailbox.CreateDB()
	os.Create("test.js")
	file, err := os.OpenFile("test.js", os.O_APPEND|os.O_WRONLY, 0644)
	file.WriteString("console.log('test');")
	file.Close()

	mb, err := mailbox.Create("test.test")
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	key := mailbox.AccessKey{FullAccess: true}
	err = key.Create()
	if err != nil {
		t.Fatal(err)
	}

	viper.Set("host", ":5112")
	viper.Set("mailbox", mb.Id)
	viper.Set("access_key", key.Secret)
	viper.Set("access_key_name", key.Name)
	viper.Set("show_requests", true)

	go serverCmd.Run(serverCmd, []string{})
	deployCmd.ParseFlags([]string{"-x"})
	deployCmd.Run(deployCmd, []string{"test.js", "test.test"})

	os.Remove("test.js")

	msg, err := mb.GetMessage()
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("No message waiting")
	}

}
