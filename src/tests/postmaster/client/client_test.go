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
var token *mailbox.AccessToken

func TestMain(m *testing.M) {
	// open database in memory for testing
	mailbox.OpenMemDB()
	mailbox.CreateDB()

	// create a default mailbox to use
	mb, _ = mailbox.Create("mb")

	// create an access token for the default mailbox
	token, _ = mailbox.CreateAPIToken("token")

	// create a postmasterClient
	pmClient = client.Client{
		Host:    "localhost:4111",
		Mailbox: mb.Id,
		Token:   token.Token,
	}

	// Start up a test server to use
	server.EnableLongPolling = false
	go server.Start(":4111")
	retCode := m.Run()

	// cleanup
	mailbox.CloseDB()
	os.Exit(retCode)
}

// TestClientGet checks to make sure the client is capable of retrieving
// messages.
func TestClientGet(t *testing.T) {
	mb.PutMessage("TEST MESSAGE")
	msg, err := pmClient.Get()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Body != msg.Body {
		t.Fatal("Message body missmatch", msg.Body)
	}
}

// TestClientPut checks to make sure the client is capable of sending messages
// to a given mailbox.
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

// TestClientDelete checks to make sure the client is capable of deleting
// messages.
func TestClientDelete(t *testing.T) {
	mb, _ := mailbox.Create("delete")
	msg, _ := mb.PutMessage("TEST DELETE")
	pmClient.Delete(msg.Id)
	count, _ := mb.MessageCount()
	if count != 0 {
		t.Fatal("Message count is", count)
	}
}

// TestNoMessage checks that a mailbox should respond with an empty response if
// there are no messages in the queue.
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

func TestSystemStats(t *testing.T) {
	mb, _ = mailbox.Create("stats.system")
	mb.PutMessage("test")
	resp, err := pmClient.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if resp.PendingMessages == 0 {
		t.Fatalf("Pending message count should not be 0")
	}
}

// BecnhMarkClientGet measures clients retrieving messages.
func BenchmarkClientGet(b *testing.B) {
	mb.PutMessage("TEST MESSAGE")
	for i := 0; i < b.N; i++ {
		pmClient.Get()
	}
}
