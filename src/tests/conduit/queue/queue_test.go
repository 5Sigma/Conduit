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
	var err error
	mb, err = mailbox.Create("mb")
	if err != nil {
		panic(err)
	}

	q = queue.New("localhost:4111", mb.Id)

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

	script, err := q.Get()
	if err != nil {
		t.Fatal(err)
	}

	if string(script.ScriptBody) != msg.Body {
		t.Fatal("Script body didnt match:", script.ScriptBody)
	}
}
