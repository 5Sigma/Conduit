package cmd

import (
	"github.com/5sigma/conduit/postmaster/client"
	"github.com/spf13/viper"
)

func MailboxClientFromConfig() (client.Client, error) {
	c := client.Client{
		Host:          viper.GetString("host"),
		AccessKeyName: viper.GetString("mailbox.name"),
		AccessKey:     viper.GetString("mailbox.key"),
		ShowRequests:  viper.GetBool("show_requests"),
		Mailbox:       viper.GetString("mailbox.name"),
	}
	return c, nil
}

func AdminClientFromConfig() (client.Client, error) {
	c := client.Client{
		Host:          viper.GetString("host"),
		AccessKeyName: viper.GetString("admin.name"),
		AccessKey:     viper.GetString("admin.key"),
		ShowRequests:  viper.GetBool("show_requests"),
		Mailbox:       viper.GetString("mailbox"),
	}
	return c, nil
}
