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

type Client struct {
	Host    string
	Token   string
	Mailbox string
}

func (client *Client) request(endpoint string, req interface{},
	res interface{}) error {
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

func (client *Client) Get() (*api.GetMessageResponse, error) {
	request := api.GetMessageRequest{
		Mailbox: client.Mailbox,
	}
	var response api.GetMessageResponse
	err := client.request("get", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (client *Client) Put(mbxs []string,
	msg string) (*api.PutMessageResponse, error) {
	request := api.PutMessageRequest{
		Mailboxes: mbxs,
		Body:      msg,
	}
	var response api.PutMessageResponse
	err := client.request("put", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (client *Client) Delete(msgId string) (*api.DeleteMessageResponse, error) {
	request := api.DeleteMessageRequest{Message: msgId}
	var response api.DeleteMessageResponse
	err := client.request("delete", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
