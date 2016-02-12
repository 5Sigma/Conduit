package mailbox

import (
	"postmaster/mailbox"
	"testing"
)

func TestCreateMailboxKey(t *testing.T) {
	mb, _ := mailbox.Create("tokentest.mailboxToken")
	key := &mailbox.AccessKey{MailboxId: mb.Id}
	err := key.Create()
	if err != nil {
		t.Fatal(err)
	}
	if !key.CanGet(mb) {
		t.Fatal("Mailbox token not able to retrieve messages.")
	}
}

func TestCreateAPIToken(t *testing.T) {
	mb, _ := mailbox.Create("tokentest.admin")
	key := mailbox.AccessKey{FullAccess: true}
	err := key.Create()
	if err != nil {
		t.Fatal(err)
	}
	if !key.CanPut(mb) {
		t.Fatal("Admin token not able to put messages.")
	}
}

func TestCreateDupKey(t *testing.T) {
	mailbox.Create("dupkeytest")
	mbKey := &mailbox.AccessKey{MailboxId: "dupkeytest"}
	mbKey.Create()
	key := &mailbox.AccessKey{Name: "dupkeytest", FullAccess: true}
	err := key.Create()
	if err == nil {
		t.Fatal("Should not create key with the same name as a mailbox")
	}
}

func TestCreateMailboxWithKeyName(t *testing.T) {
	key := &mailbox.AccessKey{Name: "dupmbtest", FullAccess: true}
	key.Create()
	_, err := mailbox.Create("dupmbtest")
	if err != nil {
		t.Fatal("Should not create mailbox with key name")
	}
}

func TestMailboxKeyName(t *testing.T) {
	mb, _ := mailbox.Create("tokentest.admin")
	key := mailbox.AccessKey{MailboxId: mb.Id}
	key.Create()
	if key.Name != mb.Id {
		t.Fatal("Key name should be set to mailbox name")
	}
}
