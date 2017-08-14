package cmd

import (
	"github.com/5sigma/conduit/postmaster/mailbox"
	"testing"
)

func TestAcessCmd(t *testing.T) {
	accessCmd.Run(accessCmd, []string{"NAME"})
	key, err := mailbox.FindKeyByName("NAME")
	if err != nil {
		t.Fatal(err)
	}
	if key == nil {
		t.Fatal("Key not created")
	}
}
