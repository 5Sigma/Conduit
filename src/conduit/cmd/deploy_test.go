package cmd

import (
	"github.com/spf13/viper"
	"io/ioutil"
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

func TestAsset(t *testing.T) {

	mailbox.OpenMemDB()
	mailbox.CreateDB()

	mb, err := mailbox.Create("test.asset")
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
	client, err := ClientFromConfig()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.CheckRemoteFile("324e63777cae0113d708633836c9cb18")
	if err != nil {
		t.Fatal(err)
	}

	if resp != false {
		t.Fatal("Check file came back true before upload.")
	}

	os.Create("test.js")
	file, err := os.OpenFile("test.js", os.O_APPEND|os.O_WRONLY, 0644)
	file.WriteString("console.log('test');")
	file.Close()

	go serverCmd.Run(serverCmd, []string{})
	deployCmd.ParseFlags([]string{"-x", "-a", "test.js"})
	deployCmd.Run(deployCmd, []string{"test.js", "test.asset"})

	defer os.Remove("test.js")

	resp, err = client.CheckRemoteFile("324e63777cae0113d708633836c9cb18")
	if err != nil {
		t.Fatal(err)
	}

	if resp != true {
		t.Fatal("File check came back false after upload")
	}

	fname, err := client.DownloadAsset("324e63777cae0113d708633836c9cb18")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(fname)

	fData, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}
	if string(fData) != "console.log('test');" {
		t.Fatal("File data is wrong")
	}
}
