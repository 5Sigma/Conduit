package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"postmaster/api"
)

// Client is used to connect to the postmaster server to receive and send
// messages through the queue.
type Client struct {
	Host    string
	Token   string
	Mailbox string
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
func (client *Client) Put(mbxs []string, pattern string,
	msg string) (*api.PutMessageResponse, error) {
	request := api.PutMessageRequest{
		Mailboxes: mbxs,
		Body:      msg,
		Pattern:   pattern,
		Token:     client.Token,
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
