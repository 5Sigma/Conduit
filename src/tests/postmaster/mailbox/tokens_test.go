package mailbox

import (
	"postmaster/mailbox"
	"testing"
)

func TestCreateMailboxToken(t *testing.T) {
	mb, _ := mailbox.Create("tokentest.mailboxToken")
	token, err := mb.CreateToken()
	if err != nil {
		t.Fatal(err)
	}
	if !mailbox.TokenCanGet(token.Token, mb.Id) {
		t.Fatal("Mailbox token not able to retrieve messages.")
	}
}

func TestCreateAPIToken(t *testing.T) {
	mb, _ := mailbox.Create("tokentest.admin")
	token, err := mailbox.CreateAPIToken("api")
	if err != nil {
		t.Fatal(err)
	}
	if !mailbox.TokenCanPut(token.Token, mb.Id) {
		t.Fatal("Admin token not able to put messages.")
	}
}
