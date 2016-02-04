package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"postmaster/api"
	"time"
)

// Client is used to connect to the postmaster server to receive and send
// messages through the queue.
type Client struct {
	Host         string
	Token        string
	Mailbox      string
	ShowRequests bool
}

// request wraps HTTP requests to the postmaster server. It is used internally
// by other functions to make various API requests. It uses a request object and
// a pointer to an empty repsonse object from the api package.
func (client *Client) request(endpoint string, req interface{},
	res interface{}) error {
	http.DefaultClient.Timeout = 0
	requestBytes, err := json.Marshal(req)
	url := fmt.Sprintf("http://%s/%s", client.Host, endpoint)
	reader := bytes.NewReader(requestBytes)
	resp, err := http.Post(url, "application/json", reader)
	if err != nil {
		return err
	}
	responseData, _ := ioutil.ReadAll(resp.Body)
	if client.ShowRequests {
		fmt.Printf("Request: %s\nResponse: %s\n", string(requestBytes),
			string(responseData))
	}
	if resp.StatusCode == 404 {
		return errors.New("API endpoint not found")
	}
	if resp.StatusCode != 200 {
		var errorResponse api.ApiError
		json.Unmarshal(responseData, &errorResponse)
		return errors.New(errorResponse.Error)
	}
	return json.Unmarshal(responseData, &res)
}

// Get retrieves a message from the server via a JSON api.
func (client *Client) Get() (*api.GetMessageResponse, error) {
	request := api.GetMessageRequest{
		Mailbox: client.Mailbox,
		Token:   client.Token,
	}

	var response api.GetMessageResponse
	err := client.request("get", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Put sends a message to a series of mailboxes. An array of mailboxes can be
// provided, as well as a pattern using '*' as wildcards. The message will by
// sent to all matching mailboxes.
func (client *Client) Put(mbxs []string, pattern string, msg string,
	deploymentName string) (*api.PutMessageResponse, error) {
	request := api.PutMessageRequest{
		Mailboxes:      mbxs,
		Body:           msg,
		Pattern:        pattern,
		Token:          client.Token,
		DeploymentName: deploymentName,
	}
	var response api.PutMessageResponse
	err := client.request("put", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Delete removes a message from the server. This is generally called after a
// message has been successfully processed to remove it from the mailbox queue.
func (client *Client) Delete(msgId string) (*api.DeleteMessageResponse, error) {
	request := api.DeleteMessageRequest{Message: msgId, Token: client.Token}
	var response api.DeleteMessageResponse
	err := client.request("delete", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (client *Client) Stats() (*api.SystemStatsResponse, error) {
	request := api.SimpleRequest{Token: client.Token}
	var response api.SystemStatsResponse
	err := client.request("stats", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (client *Client) ListDeploys(namePattern string, limitToken bool,
	count int) (*api.DeploymentStatsResponse, error) {
	request := api.DeploymentStatsRequest{
		Token:        client.Token,
		Count:        int64(count),
		NamePattern:  namePattern,
		TokenPattern: ".*",
	}
	if limitToken {
		request.TokenPattern = client.Token
	}
	var response api.DeploymentStatsResponse
	err := client.request("deploy/list", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (client *Client) DeploymentDetail(id string) (*api.DeploymentStatsResponse, error) {
	request := api.DeploymentStatsRequest{Token: client.Token, Deployment: id}
	var response api.DeploymentStatsResponse
	err := client.request("deploy/list", request, &response)
	return &response, err
}

func (client *Client) Respond(messageId string, msg string) error {
	request := api.ResponseRequest{
		Token:    client.Token,
		Response: msg,
		Message:  messageId,
	}
	var response api.SimpleResponse
	err := client.request("deploy/respond", request, &response)
	return err
}

func (client *Client) PollDeployment(depId string,
	f func(*api.DeploymentStats) bool) (*api.DeploymentStats, error) {
	request := api.DeploymentStatsRequest{Token: client.Token, Deployment: depId}
	var response *api.DeploymentStatsResponse
	loop := true
	for loop != false {
		err := client.request("deploy/list", request, &response)
		if err != nil {
			return nil, err
		}
		if len(response.Deployments) == 0 {
			return nil, errors.New("Could not find deployment")
		}
		loop = f(&response.Deployments[0])
		time.Sleep(1 * time.Second)
	}
	return &response.Deployments[0], nil
}

func (client *Client) RegisterMailbox(m string) (*api.RegisterResponse, error) {
	request := &api.RegisterRequest{
		Token:   client.Token,
		Mailbox: m,
	}
	var response *api.RegisterResponse
	err := client.request("register", request, &response)
	return response, err
}
