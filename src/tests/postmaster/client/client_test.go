package client

import (
	"conduit/log"
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
	log.LogStdOut = false
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
	_, err := pmClient.Put([]string{mb1.Id, mb2.Id}, "", "PUT TEST", "")
	if err != nil {
		t.Fatal(err)
	}
	count1, _ := mb1.MessageCount()
	count2, _ := mb2.MessageCount()
	if count1 != 1 || count2 != 1 {
		t.Fatal("Message counts are ", count1, ",", count2)
	}
}

func TestAutoCreateDeploy(t *testing.T) {
	mb, _ := mailbox.Create("put.autocreate.deploy")
	msg, err := pmClient.Put([]string{mb.Id}, "", "TEST MESSAGE", "blah")
	if err != nil {
		t.Fatal(err)
	}
	if msg.Deployment == "" {
		t.Fatal("Deployment is empty")
	}
	dep, err := mailbox.FindDeployment(msg.Deployment)
	if err != nil {
		t.Fatal(err)
	}
	if dep.Name != "blah" {
		t.Fatal("Deployment name not set")
	}
}

func TestResponse(t *testing.T) {
	mb, err := mailbox.Create("deployment.response")
	if err != nil {
		t.Fatal(err)
	}
	dep := &mailbox.Deployment{
		Name:        "dep",
		DeployedBy:  token.Token,
		MessageBody: "testMessage",
	}
	err = dep.Create()
	if err != nil {
		t.Fatal(err)
	}
	msg, err := mb.DeployMessage(dep.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = pmClient.Respond(msg.Id, "testing repsonse")
	if err != nil {
		t.Fatal(err)
	}

	responses, err := dep.GetResponses()
	if err != nil {
		t.Fatal(err)
	}

	if len(responses) == 0 {
		t.Fatal("Response was not added")
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
		t.Fatal("Pending message count should not be 0")
	}
}

func TestListDeploys(t *testing.T) {
	mb, err := mailbox.Create("deployment.list")
	if err != nil {
		t.Fatal(err)
	}
	dep := &mailbox.Deployment{
		Name:        "dep",
		DeployedBy:  token.Token,
		MessageBody: "test message",
	}
	err = dep.Create()
	if err != nil {
		t.Fatal(err)
	}
	_, err = mb.DeployMessage(dep.Id)
	if err != nil {
		t.Fatal(err)
	}
	pmClient.Mailbox = mb.Id
	resp, err := pmClient.ListDeploys(".*", false, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Deployments) == 0 {
		t.Fatal("No deployments returned")
	}
}

func TestDeploymentDetail(t *testing.T) {
	mb, err := mailbox.Create("deployment.detail")
	if err != nil {
		t.Fatal(err)
	}
	dep := mailbox.Deployment{MessageBody: "test message"}
	err = dep.Create()
	if err != nil {
		t.Fatal(err)
	}
	_, err = mb.DeployMessage(dep.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = dep.AddResponse(mb.Id, "test repsonse")
	if err != nil {
		t.Fatal(err)
	}
	pmClient.Mailbox = mb.Id
	resp, err := pmClient.DeploymentDetail(dep.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Deployments) == 0 {
		t.Fatal("No deployments returned")
	}
	if len(resp.Deployments[0].Responses) == 0 {
		t.Fatal("No deployment responses returned")
	}
}

// BecnhMarkClientGet measures clients retrieving messages.
func BenchmarkClientGet(b *testing.B) {
	mb.PutMessage("TEST MESSAGE")
	for i := 0; i < b.N; i++ {
		pmClient.Get()
	}
}
