package client

import (
	"os"
	"postmaster/client"
	"postmaster/mailbox"
	"postmaster/server"
	"testing"
)

var pmClient client.Client
var mb *mailbox.Mailbox

func TestMain(m *testing.M) {
	mailbox.OpenMemDB()
	mailbox.CreateDB()
	mb, _ = mailbox.Create("mb")
	pmClient = client.Client{
		Host:    "localhost:4111",
		Mailbox: mb.Id,
	}
	go server.Start(":4111")
	retCode := m.Run()

	mailbox.CloseDB()
	os.Exit(retCode)
}

func TestClientGet(t *testing.T) {
	mb.PutMessage("TEST MESSAGE")
	msg, err := pmClient.Get()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Body != "TEST MESSAGE" {
		t.Fatal("Message body missmatch", msg.Body)
	}
}

func TestClientPut(t *testing.T) {
	mb1, _ := mailbox.Create("put1")
	mb2, _ := mailbox.Create("put2")
	_, err := pmClient.Put([]string{mb1.Id, mb2.Id}, "", "PUT TEST")
	if err != nil {
		t.Fatal(err)
	}
	count1, _ := mb1.MessageCount()
	count2, _ := mb2.MessageCount()
	if count1 != 1 || count2 != 1 {
		t.Fatal("Message counts are ", count1, ",", count2)
	}
}

func TestClientDelete(t *testing.T) {
	mb, _ := mailbox.Create("delete")
	msg, _ := mb.PutMessage("TEST DELETE")
	pmClient.Delete(msg.Id)
	count, _ := mb.MessageCount()
	if count != 0 {
		t.Fatal("Message count is", count)
	}
}

func TestNoMessages(t *testing.T) {
	mb, _ = mailbox.Create("empty")
	pmClient.Mailbox = mb.Id
	resp, err := pmClient.Get()
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsEmpty() {
		t.Fatal("Response body is not empty")
	}
}

func BenchmarkClientGet(b *testing.B) {
	mb.PutMessage("TEST MESSAGE")
	for i := 0; i < b.N; i++ {
		pmClient.Get()
	}
}
