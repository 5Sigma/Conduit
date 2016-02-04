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

type TestDeployment struct {
	Token      *mailbox.AccessToken
	Deployment *mailbox.Deployment
	Mailbox    *mailbox.Mailbox
	Message    *mailbox.Message
}

func generateDeployment() (*TestDeployment, error) {
	token, err := mailbox.CreateAPIToken(mailbox.GenerateIdentifier())
	if err != nil {
		return nil, err
	}
	deployment := &mailbox.Deployment{
		MessageBody: mailbox.GenerateIdentifier(),
		Name:        mailbox.GenerateIdentifier(),
		DeployedBy:  token.Token,
	}
	err = deployment.Create()
	if err != nil {
		return nil, err
	}
	mb, err := mailbox.Create(mailbox.GenerateIdentifier())
	if err != nil {
		return nil, err
	}
	msg, err := deployment.Deploy(mb)
	if err != nil {
		return nil, err
	}
	return &TestDeployment{
		Token:      token,
		Deployment: deployment,
		Mailbox:    mb,
		Message:    msg,
	}, nil
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

func TestGet(t *testing.T) {
	dep, _ := generateDeployment()
	req := api.GetMessageRequest{Mailbox: dep.Mailbox.Id, Token: dep.Token.Token}
	var resp api.GetMessageResponse
	doRequest(t, req, &resp, "get")
	if resp.Body != dep.Message.Body {
		t.Fatalf("Message body TEST!=%s", dep.Message.Body)
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
	if message1 == nil || message2 == nil {
		t.Fatal("Message is nil")
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
	token, _ := mb.CreateToken()
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

func TestDeployInfoList(t *testing.T) {
	mb, _ := mailbox.Create("stats.deployinfo")
	token, err := mailbox.CreateAPIToken("stats.deployinfo")
	mb.PutMessage("test")
	mb.PutMessage("test2")
	if err != nil {
		t.Fatal(err)
	}

	req := api.DeploymentStatsRequest{
		Token: token.Token,
		Count: 2,
	}
	var resp api.DeploymentStatsResponse
	code := doRequest(t, req, &resp, "deploy/list")
	if code != 200 {
		t.Fatalf("Server repsponded with %d", code)
	}
	if len(resp.Deployments) != 2 {
		t.Fatalf("Deployment count %d!=2", len(resp.Deployments))
	}
}

func TestDeployInfoListByToken(t *testing.T) {
	dep1, err := generateDeployment()
	if err != nil {
		t.Fatal(err)
	}
	dep2, err := generateDeployment()
	if err != nil {
		t.Fatal(err)
	}
	req := api.DeploymentStatsRequest{
		Token:        dep1.Token.Token,
		Count:        2,
		TokenPattern: dep2.Token.Token,
	}
	var resp api.DeploymentStatsResponse
	code := doRequest(t, req, &resp, "deploy/list")
	if code != 200 {
		t.Fatalf("Server repsponded with %d", code)
	}
	if len(resp.Deployments) != 1 {
		t.Fatalf("Deployment count %d!=1", len(resp.Deployments))
	}
}

func TestDeployListByName(t *testing.T) {
	deployment1, err := generateDeployment()
	if err != nil {
		t.Fatal(err)
	}
	deployment1.Deployment.Name = "test"
	deployment1.Deployment.Save()
	generateDeployment()

	if err != nil {
		t.Fatal(err)
	}

	req := api.DeploymentStatsRequest{
		Token:       deployment1.Token.Token,
		Count:       10,
		NamePattern: "t*t",
	}

	var resp api.DeploymentStatsResponse
	code := doRequest(t, req, &resp, "deploy/list")
	if code != 200 {
		t.Fatalf("Server repsponded with %d", code)
	}
	if len(resp.Deployments) != 1 {
		t.Fatalf("Deployment length %d != 1", len(resp.Deployments))
	}
}

func TestRegister(t *testing.T) {
	token, err := mailbox.CreateAPIToken("tesT")
	if err != nil {
		t.Fatal(err)
	}
	req := api.RegisterRequest{
		Token:   token.Token,
		Mailbox: "register.test",
	}
	var resp api.RegisterResponse
	code := doRequest(t, req, &resp, "register")
	if code != 200 {
		t.Fatalf("Server repsponded with %d", code)
	}
	mb, err := mailbox.Find("register.test")
	if err != nil {
		t.Fatal(err)
	}
	if mb == nil {
		t.Fatal("Mailbox not registered")
	}
	if !mailbox.TokenCanGet(token.Token, mb.Id) {
		t.Fatal("Token not bound to mailbox")
	}
}
