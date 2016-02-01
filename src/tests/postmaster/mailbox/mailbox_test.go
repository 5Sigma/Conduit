package mailbox

import (
	"os"
	"postmaster/mailbox"
	"testing"
)

func TestMain(m *testing.M) {
	mailbox.OpenMemDB()
	mailbox.CreateDB()

	retCode := m.Run()

	mailbox.CloseDB()
	os.Exit(retCode)
}

func TestSearch(t *testing.T) {
	_, err := mailbox.Create("org.Test")
	if err != nil {
		t.Fatal(err)
	}
	results, err := mailbox.Search("Org.*")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("Mailbox not found.")
	}
}
