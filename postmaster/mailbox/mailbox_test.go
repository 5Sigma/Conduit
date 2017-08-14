package mailbox

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	OpenMemDB()
	CreateDB()

	retCode := m.Run()

	CloseDB()
	os.Exit(retCode)
}

func TestSearch(t *testing.T) {
	_, err := Create("org.Test")
	if err != nil {
		t.Fatal(err)
	}
	results, err := Search("Org.*")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("Mailbox not found.")
	}
}
