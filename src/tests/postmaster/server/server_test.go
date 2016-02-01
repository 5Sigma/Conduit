package server

import (
	"bytes"
	"conduit/log"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"postmaster/api"
	"postmaster/mailbox"
	"postmaster/server"
	"testing"
)

func TestMain(m *testing.M) {
	mailbox.OpenMemDB()
	mailbox.CreateDB()

	server.EnableLongPolling = false
	server.ThrottleDelay = 0
	log.LogStdOut = false

	go server.Start(":4111")

	retCode := m.Run()

	mailbox.CloseDB()
	os.Exit(retCode)
}

func doRequest(t *testing.T, req interface{}, response interface{},
	url string) int {
	requestData, _ := json.Marshal(&req)
	resp, err := http.Post("http://localhost:4111/"+url, "application/json",
		bytes.NewReader(requestData))
	if err != nil {
		t.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	t.Logf("\nUrl: %s\nRequest: %s\nResponse:%s\n", url, string(requestData),
		string(responseData))
	json.Unmarshal(responseData, response)
	return resp.StatusCode
}

func TestGet(t *testing.T) {
	mb, err := mailbox.Create("get")
	if err != nil {
		t.Fatal(err)
	}
	mb.PutMessage("TEST")
	token, _ := mailbox.CreateMailboxToken(mb.Id)
	req := api.GetMessageRequest{Mailbox: mb.Id, Token: token.Token}
	var resp api.GetMessageResponse
	doRequest(t, req, &resp, "get")
	if resp.Body != "TEST" {
		t.Fatal("Message body is not correct")
	}
	if resp.ReceiveCount != 1 {
		t.Fatal("Message receiveCount is not 1")
	}
	doRequest(t, req, &resp, "get")
	if resp.ReceiveCount != 2 {
		t.Fatal("Message receiveCount did not increase to 2 on second call")
	}
}

func TestGetBadToken(t *testing.T) {
	mb, err := mailbox.Create("get.badtoken")
	if err != nil {
		t.Fatal(err)
	}
	mb.PutMessage("TEST")
	req := api.GetMessageRequest{Mailbox: mb.Id, Token: "BADTOKEN"}
	var resp api.GetMessageResponse
	code := doRequest(t, req, &resp, "get")
	if code == 200 {
		t.Fatal("Bad token should respond with an error")
	}
}

func TestPut(t *testing.T) {
	mb1, _ := mailbox.Create("put1")
	mb2, _ := mailbox.Create("put2")
	token, _ := mailbox.CreateAPIToken("test")
	req := api.PutMessageRequest{
		Token: token.Token,
		Mailboxes: []string{
			mb1.Id,
			mb2.Id,
		},
		Body: "TEST",
	}
	var resp api.PutMessageResponse
	code := doRequest(t, req, &resp, "put")
	count, err := mb1.MessageCount()
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("Message not added to mailbox")
	}
	count, err = mb2.MessageCount()
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("Message not added to mailbox")
	}
	if code != 200 {
		t.Fatal("Server responded with", code)
	}
	message1, err := mb1.GetMessage()
	if err != nil {
		t.Fatal(err)
	}
	message2, err := mb2.GetMessage()
	if err != nil {
		t.Fatal(err)
	}
	if message1.Body != "TEST" {
		t.Fatal("Incorrect message1 body", message1.Body)
	}
	if message2.Body != "TEST" {
		t.Fatal("Incorrect message2 body", message2.Body)
	}
}

func TestPutBadToken(t *testing.T) {
	mb, _ := mailbox.Create("puttest.badtoken")
	token, _ := mailbox.CreateMailboxToken(mb.Id)
	req := api.PutMessageRequest{
		Mailboxes: []string{mb.Id},
		Body:      "TEST MESSAGE",
		Token:     token.Token,
	}
	var resp api.PutMessageResponse
	code := doRequest(t, req, &resp, "put")
	if code == 200 {
		t.Fatal("Bad token should return error")
	}
}

func TestPutByPattern(t *testing.T) {
	mb, _ := mailbox.Create("PATTERN")
	token, _ := mailbox.CreateAPIToken("test")
	req := api.PutMessageRequest{Pattern: "P*", Token: token.Token}
	var resp api.PutMessageResponse
	code := doRequest(t, req, &resp, "put")
	count, err := mb.MessageCount()
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("Message not added to mailbox")
	}
	if code != 200 {
		t.Fatal("Server responded with", code)
	}
}

func TestDelete(t *testing.T) {
	mb, err := mailbox.Create("delete")
	if err != nil {
		t.Fatal(err)
	}

	msg, err := mb.PutMessage("TEST")
	if err != nil {
		t.Fatal(err)
	}

	req := api.DeleteMessageRequest{Message: msg.Id}
	resp := api.DeleteMessageResponse{}

	statusCode := doRequest(t, req, &resp, "delete")
	if statusCode != 200 {
		t.Fatal("Server responded with", statusCode)
	}

	count, err := mb.MessageCount()
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatal("Message count should be 0 but is", count)
	}
}

func TestBadMailbox(t *testing.T) {
	req := api.GetMessageRequest{Mailbox: "111"}
	var resp api.GetMessageResponse
	code := doRequest(t, req, &resp, "get")
	if code != 400 {
		t.Fatal("Should of responded with 400 but it responded with", code)
	}
}
func TestSystemStats(t *testing.T) {
	mailbox.Create("stats.systemtest")
	token, err := mailbox.CreateAPIToken("systemstatstoken")
	if err != nil {
		t.Fatal(err)
	}
	req := api.SimpleRequest{Token: token.Token}
	var resp api.SystemStatsResponse
	code := doRequest(t, req, &resp, "stats")
	if code != 200 {
		t.Fatal("Server responded with", code)
	}
}
