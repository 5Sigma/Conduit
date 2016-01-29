package mailbox

import (
	"postmaster/mailbox"
	"testing"
)

func TestSearch(t *testing.T) {
	mailbox.OpenMemDB()
	mailbox.CreateDB()
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
