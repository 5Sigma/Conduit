package cmd

import (
	"postmaster/mailbox"
	"testing"
)

func TestServerPurge(t *testing.T) {
	mb, _ := mailbox.Create("purge.test")
	msg, _ := mb.PutMessage("tesT")
	if msg == nil {
		t.Fatal("Message not deployed")
	}
	mailbox.CloseDB()
	purgeCmd.Run(purgeCmd, []string{"purge.test"})
	msg, _ = mb.GetMessage()
	if msg != nil {
		t.Fatal("Message was not purged")
	}
}
