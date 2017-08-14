package mailbox

import (
	"testing"
)

func TestCreateMailboxKey(t *testing.T) {
	mb, _ := Create("tokentest.mailboxToken")
	key := &AccessKey{MailboxId: mb.Id}
	err := key.Create()
	if err != nil {
		t.Fatal(err)
	}
	if !key.CanGet(mb) {
		t.Fatal("Mailbox token not able to retrieve messages.")
	}
}

func TestCreateAPIToken(t *testing.T) {
	mb, _ := Create("tokentest.admin")
	key := AccessKey{FullAccess: true}
	err := key.Create()
	if err != nil {
		t.Fatal(err)
	}
	if !key.CanPut(mb) {
		t.Fatal("Admin token not able to put messages.")
	}
}

func TestCreateDupKey(t *testing.T) {
	Create("dupkeytest")
	mbKey := &AccessKey{MailboxId: "dupkeytest"}
	mbKey.Create()
	key := &AccessKey{Name: "dupkeytest", FullAccess: true}
	err := key.Create()
	if err == nil {
		t.Fatal("Should not create key with the same name as a mailbox")
	}
}

func TestCreateMailboxWithKeyName(t *testing.T) {
	key := &AccessKey{Name: "dupmbtest", FullAccess: true}
	key.Create()
	_, err := Create("dupmbtest")
	if err != nil {
		t.Fatal("Should not create mailbox with key name")
	}
}

func TestMailboxKeyName(t *testing.T) {
	mb, _ := Create("tokentest.admin")
	key := AccessKey{MailboxId: mb.Id}
	key.Create()
	if key.Name != mb.Id {
		t.Fatal("Key name should be set to mailbox name")
	}
}
