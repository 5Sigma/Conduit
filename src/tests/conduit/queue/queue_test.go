package queue

import (
	"conduit/queue"
	"os"
	"postmaster/mailbox"
	"postmaster/server"
	"testing"
)

var q queue.Queue
var mb *mailbox.Mailbox

func TestMain(m *testing.M) {
	mailbox.OpenMemDB()
	mailbox.CreateDB()
	mb, _ = mailbox.Create("mb")
	token, _ := mailbox.CreateMailboxToken(mb.Id)

	q = queue.New("localhost:4111", mb.Id, token.Token)

	server.EnableLongPolling = false
	go server.Start(":4111")

	retCode := m.Run()

	mailbox.CloseDB()
	os.Exit(retCode)
}

func TestQueueGet(t *testing.T) {
	msg, err := mb.PutMessage("Test script")
	if err != nil {
		t.Fatal(err)
	}

	if msg == nil {
		t.Fatal("Message did not get returned from PutMessage")
	}

	script, err := q.Get()
	if err != nil {
		t.Fatal(err)
	}

	if script == nil {
		t.Fatal("No script returned from Get")
	}

	if string(script.ScriptBody) != msg.Body {
		t.Fatal("Script body didnt match:", script.ScriptBody)
	}
}
