package server

import (
	"bytes"
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
	AccessKey  *mailbox.AccessKey
	Deployment *mailbox.Deployment
	Mailbox    *mailbox.Mailbox
	Message    *mailbox.Message
}

func generateDeployment() (*TestDeployment, error) {
	mb, err := mailbox.Create(mailbox.GenerateIdentifier())
	if err != nil {
		return nil, err
	}
	accessKey := &mailbox.AccessKey{MailboxId: mb.Id}
	err = accessKey.Create()
	if err != nil {
		return nil, err
	}
	deployment := &mailbox.Deployment{
		MessageBody: mailbox.GenerateIdentifier(),
		Name:        mailbox.GenerateIdentifier(),
		DeployedBy:  accessKey.Name,
	}
	err = deployment.Create()
	if err != nil {
		return nil, err
	}
	msg, err := deployment.Deploy(mb)
	if err != nil {
		return nil, err
	}
	return &TestDeployment{
		AccessKey:  accessKey,
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

	go server.Start(":4111")

	retCode := m.Run()

	mailbox.CloseDB()
	os.Exit(retCode)
}

func TestGet(t *testing.T) {
	dep, err := generateDeployment()
	if err != nil {
		t.Fatal(err)
	}
	req := api.GetMessageRequest{Mailbox: dep.Mailbox.Id}
	req.Sign(dep.AccessKey.Name, dep.AccessKey.Secret)
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

func TestGetBadSignature(t *testing.T) {
	mb, err := mailbox.Create("get.badtoken")
	if err != nil {
		t.Fatal(err)
	}
	mb.PutMessage("TEST")
	req := api.GetMessageRequest{Mailbox: mb.Id}
	var resp api.GetMessageResponse
	code := doRequest(t, req, &resp, "get")
	if code == 200 {
		t.Fatal("Bad token should respond with an error")
	}
}

func TestPut(t *testing.T) {
	mb1, err := mailbox.Create("put1")
	if err != nil {
		t.Fatal(err)
	}
	mb2, err := mailbox.Create("put2")
	if err != nil {
		t.Fatal(err)
	}
	accessKey := mailbox.AccessKey{FullAccess: true}
	accessKey.Create()
	req := api.PutMessageRequest{
		Mailboxes: []string{
			mb1.Id,
			mb2.Id,
		},
		Body: "TEST",
	}
	req.Sign(accessKey.Name, accessKey.Secret)
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
	accessKey := mailbox.AccessKey{MailboxId: mb.Id}
	accessKey.Create()
	req := api.PutMessageRequest{
		Mailboxes: []string{mb.Id},
		Body:      "TEST MESSAGE",
	}
	req.Sign(accessKey.Name, accessKey.Secret)
	var resp api.PutMessageResponse
	code := doRequest(t, req, &resp, "put")
	if code == 200 {
		t.Fatal("Bad token should return error")
	}
}

func TestPutByPattern(t *testing.T) {
	mb, _ := mailbox.Create("PATTERN")
	accessKey := mailbox.AccessKey{FullAccess: true}
	accessKey.Create()
	req := api.PutMessageRequest{Pattern: "P*"}
	req.Sign(accessKey.Name, accessKey.Secret)
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

	key := mailbox.AccessKey{MailboxId: mb.Id}
	key.Create()
	req := api.DeleteMessageRequest{Message: msg.Id}
	req.Sign(key.Name, key.Secret)
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
	key := mailbox.AccessKey{FullAccess: true}
	key.Create()
	req := api.GetMessageRequest{Mailbox: "111"}
	req.Sign(key.Name, key.Secret)
	var resp api.GetMessageResponse
	code := doRequest(t, req, &resp, "get")
	if code != 400 {
		t.Fatal("Should of responded with 400 but it responded with", code)
	}
}

func TestSystemStats(t *testing.T) {
	mailbox.Create("stats.systemtest")
	req := api.SimpleRequest{}
	accessKey := mailbox.AccessKey{FullAccess: true}
	accessKey.Create()
	req.Sign(accessKey.Name, accessKey.Secret)
	var resp api.SystemStatsResponse
	code := doRequest(t, req, &resp, "stats")
	if code != 200 {
		t.Fatal("Server responded with", code)
	}
}

func TestDeployInfoList(t *testing.T) {
	mb, _ := mailbox.Create("stats.deployinfo")
	mb.PutMessage("test")
	mb.PutMessage("test2")
	key := mailbox.AccessKey{FullAccess: true}
	key.Create()
	req := api.DeploymentStatsRequest{Count: 2}
	req.Sign(key.Name, key.Secret)
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
	_, err := generateDeployment()
	key := mailbox.AccessKey{FullAccess: true}
	key.Create()
	if err != nil {
		t.Fatal(err)
	}
	dep2, err := generateDeployment()
	if err != nil {
		t.Fatal(err)
	}
	req := api.DeploymentStatsRequest{
		Count:        2,
		TokenPattern: dep2.AccessKey.Name,
	}
	req.Sign(key.Name, key.Secret)
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
	key := mailbox.AccessKey{FullAccess: true}
	key.Create()
	req := api.DeploymentStatsRequest{
		Count:       10,
		NamePattern: "t*t",
	}
	req.Sign(key.Name, key.Secret)

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
	key := mailbox.AccessKey{FullAccess: true}
	key.Create()
	req := api.RegisterRequest{Mailbox: "register.test"}
	req.Sign(key.Name, key.Secret)
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
	if !key.CanGet(mb) {
		t.Fatal("Key not bound to mailbox")
	}
}

func TestFileUpload(t *testing.T) {
	key := mailbox.AccessKey{FullAccess: true}
	key.Create()

}
