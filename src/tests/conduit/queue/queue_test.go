package queue

import (
	"conduit/queue"
	"os"
	"postmaster/mailbox"
	"postmaster/server"
	"testing"
	"time"
)

var q queue.Queue
var mb *mailbox.Mailbox

func TestMain(m *testing.M) {
	mailbox.CreateDB()
	var err error
	mb, err = mailbox.Create()
	if err != nil {
		panic(err)
	}

	q = queue.New("localhost:4111", mb.Id)

	go server.Start(":4111")
	time.Sleep(500)

	retCode := m.Run()

	mailbox.CloseDB()
	os.Remove("mailboxes.db")
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
