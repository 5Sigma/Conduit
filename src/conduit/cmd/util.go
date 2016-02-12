package cmd

import (
	"github.com/spf13/viper"
	"postmaster/client"
)

func ClientFromConfig() (client.Client, error) {
	c := client.Client{
		Host:          viper.GetString("host"),
		AccessKeyName: viper.GetString("access_key_name"),
		AccessKey:     viper.GetString("access_key"),
		ShowRequests:  viper.GetBool("show_requests"),
		Mailbox:       viper.GetString("mailbox"),
	}
	if c.AccessKeyName == "" {
		c.AccessKeyName = viper.GetString("mailbox")
	}
	return c, nil
}
