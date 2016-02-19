package client

import (
	"bytes"
	"conduit/log"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"postmaster/api"
	"time"
)

// Client is used to connect to the postmaster server to receive and send
// messages through the queue.
type Client struct {
	Host          string
	AccessKeyName string
	AccessKey     string
	Mailbox       string
	ShowRequests  bool
	UseProxy      bool
	ProxyAddress  string
}

func (client *Client) getHttpClient() (*http.Client, error) {
	var hClient *http.Client
	if client.UseProxy {
		pxUrl, err := url.Parse(client.ProxyAddress)
		if err != nil {
			return nil, errors.New("Invalid proxy address")
		}
		hClient = &http.Client{
			Timeout:   0,
			Transport: &http.Transport{Proxy: http.ProxyURL(pxUrl)},
		}
	} else {
		hClient = &http.Client{Timeout: 0}
	}
	return hClient, nil
}

// request wraps HTTP requests to the postmaster server. It is used internally
// by other functions to make various API requests. It uses a request object and
// a pointer to an empty repsonse object from the api package.
func (client *Client) request(endpoint string, req interface{}, res interface{}) error {
	hClient, err := client.getHttpClient()
	if err != nil {
		return err
	}

	http.DefaultClient.Timeout = 0
	requestBytes, err := json.Marshal(req)
	url := fmt.Sprintf("http://%s/%s", client.Host, endpoint)
	reader := bytes.NewReader(requestBytes)
	resp, err := hClient.Post(url, "application/json", reader)
	if err != nil {
		return err
	}
	responseData, _ := ioutil.ReadAll(resp.Body)
	if client.ShowRequests == true {
		fmt.Printf("URL: %s\nRequest: %s\nResponse: %s\n", url, string(requestBytes),
			string(responseData))
	}
	if resp.StatusCode == 404 {
		return errors.New("API endpoint not found")
	}
	if resp.StatusCode == 426 {
		log.Warn("Server version does not match")
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
	request := api.GetMessageRequest{Mailbox: client.Mailbox}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response api.GetMessageResponse
	err := client.request("get", request, &response)
	if err != nil {
		return nil, err
	}
	if !response.IsEmpty() {
		if !response.Validate(client.AccessKey) {
			return nil, errors.New("Could not validate the server's signature")
		}
	}
	return &response, nil
}

// Put sends a message to a series of mailboxes. An array of mailboxes can be
// provided, as well as a pattern using '*' as wildcards. The message will by
// sent to all matching mailboxes.
func (client *Client) Put(mbxs []string, pattern string, msg string,
	deploymentName string, asset string) (*api.PutMessageResponse, error) {
	md5, _ := client.hashFile(asset)
	if asset != "" {
		exists, err := client.CheckRemoteFile(md5)
		if err != nil {
			return nil, err
		}
		if exists == true {
			log.Info("File exists on server, skipping upload")
		} else {
			_, err := client.Upload(asset)
			if err != nil {
				log.Debug(err.Error())
				return nil, errors.New("Could not upload asset")
			}
		}
	}

	request := api.PutMessageRequest{
		Mailboxes:      mbxs,
		Body:           msg,
		Pattern:        pattern,
		DeploymentName: deploymentName,
		Asset:          md5,
	}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response api.PutMessageResponse
	err := client.request("put", request, &response)
	if err != nil {
		return nil, err
	}
	if !response.Validate(client.AccessKey) {
		return nil, errors.New("Server responded with an invalid signature")
	}
	return &response, nil
}

// Delete removes a message from the server. This is generally called after a
// message has been successfully processed to remove it from the mailbox queue.
func (client *Client) Delete(msgId string) (*api.DeleteMessageResponse, error) {
	request := api.DeleteMessageRequest{Message: msgId}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response api.DeleteMessageResponse
	err := client.request("delete", request, &response)
	if err != nil {
		return nil, err
	}
	if !response.Validate(client.AccessKey) {
		return nil, errors.New("Could not validate signature")
	}
	return &response, nil
}

func (client *Client) Stats() (*api.SystemStatsResponse, error) {
	request := api.SimpleRequest{}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response api.SystemStatsResponse
	err := client.request("stats", request, &response)
	if err != nil {
		return nil, err
	}
	if !response.Validate(client.AccessKey) {
		return nil, errors.New("Could not validate signature")
	}
	return &response, nil
}

func (client *Client) ClientStatus() ([]api.ClientStatus, error) {
	request := api.SimpleRequest{}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response api.ClientStatusResponse
	err := client.request("stats/clients", request, &response)
	if err != nil {
		return nil, err
	}
	if !response.Validate(client.AccessKey) {
		return nil, errors.New("Could not validate signature")
	}
	return response.Clients, nil
}

func (client *Client) ListDeploys(namePattern string, limitToken bool,
	count int) (*api.DeploymentStatsResponse, error) {
	request := api.DeploymentStatsRequest{
		Count:        int64(count),
		NamePattern:  namePattern,
		TokenPattern: ".*",
	}
	request.Sign(client.AccessKeyName, client.AccessKey)
	if limitToken {
		request.TokenPattern = client.AccessKeyName
	}
	var response api.DeploymentStatsResponse
	err := client.request("deploy/list", request, &response)
	if err != nil {
		return nil, err
	}
	if !response.Validate(client.AccessKey) {
		return nil, errors.New("Could not validate signature")
	}
	return &response, nil
}

func (client *Client) DeploymentDetail(id string) (*api.DeploymentStatsResponse, error) {
	request := api.DeploymentStatsRequest{Deployment: id}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response api.DeploymentStatsResponse
	err := client.request("deploy/list", request, &response)
	if !response.Validate(client.AccessKey) {
		return nil, errors.New("Could not validate signature")
	}
	return &response, err
}

func (client *Client) Respond(messageId string, msg string, isErr bool) error {
	request := api.ResponseRequest{
		Response: msg,
		Message:  messageId,
		Error:    isErr,
	}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response api.SimpleResponse
	err := client.request("deploy/respond", request, &response)
	if err != nil {
		return err
	}
	if !response.Validate(client.AccessKey) {
		return errors.New("Could not validate signature")
	}
	return nil
}

func (client *Client) PollDeployment(depId string,
	f func(*api.DeploymentStats) bool) (*api.DeploymentStats, error) {
	loop := true
	var response *api.DeploymentStatsResponse
	for loop != false {
		request := api.DeploymentStatsRequest{Deployment: depId}
		request.Sign(client.AccessKeyName, client.AccessKey)
		err := client.request("deploy/list", request, &response)
		if err != nil {
			return nil, err
		}
		if !response.Validate(client.AccessKey) {
			return nil, errors.New("Could not validate signature")
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
	request := &api.RegisterRequest{Mailbox: m}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response *api.RegisterResponse
	err := client.request("register", request, &response)
	if err != nil {
		return nil, err
	}
	if !response.Validate(client.AccessKey) {
		return nil, errors.New("Could not validate signature")
	}
	return response, err
}

func (client *Client) DeregisterMailbox(m string) (*api.SimpleResponse, error) {
	request := &api.RegisterRequest{Mailbox: m}
	request.Sign(client.AccessKeyName, client.AccessKey)
	var response *api.SimpleResponse
	err := client.request("deregister", request, &response)
	if err != nil {
		return nil, err
	}
	if !response.Validate(client.AccessKey) {
		return nil, errors.New("Could not validate signature")
	}
	return response, err
}

func (client *Client) hashFile(fp string) (string, error) {
	file, err := os.Open(fp)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		return "", err
	}
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashStr := fmt.Sprintf("%x", hash.Sum(nil))
	return hashStr, nil
}

func (client *Client) Upload(fpath string) (*api.SimpleResponse, error) {

	var (
		body   = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
		req    = api.UploadFileRequest{Filename: filepath.Base(fpath)}
		err    error
	)

	defer writer.Close()

	req.MD5, err = client.hashFile(fpath)
	if err != nil {
		return nil, err
	}

	log.Infof("Uploading file %s", fpath)

	hClient, err := client.getHttpClient()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/%s", client.Host, "upload")

	// Write JSON data as post param
	req.Sign(client.AccessKeyName, client.AccessKey)
	requestBytes, err := json.Marshal(req)

	err = writer.WriteField("data", string(requestBytes))
	if err != nil {
		return nil, err
	}

	// Write file to request
	file, err := os.Open(fpath)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		return nil, err
	}
	part, err := writer.CreateFormFile("file", fpath)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	writer.Close()

	// Make the request
	hReq, err := http.NewRequest("POST", url, body)
	hReq.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := hClient.Do(hReq)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	// Read the response
	responseData, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var errorResponse api.ApiError
		json.Unmarshal(responseData, &errorResponse)
		return nil, errors.New(errorResponse.Error)
	}
	res := &api.SimpleResponse{}
	json.Unmarshal(responseData, res)
	return res, err
}

// Checks if the remote server has a specified file. Before upload a file this
// method is called to verify that the server does not already contian this
// file. That way clients do not upload the file when unnessecary.
func (client *Client) CheckRemoteFile(md5 string) (bool, error) {
	req := &api.CheckFileRequest{MD5: md5}
	req.Sign(client.AccessKeyName, client.AccessKey)
	resp := api.SimpleResponse{}
	err := client.request("checkfile", req, &resp)
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func (client *Client) DownloadAsset(md5 string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "conduit.")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	req := api.GetAssetRequest{
		MD5: md5,
	}
	req.Sign(client.AccessKeyName, client.AccessKey)

	requestBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("http://%s/%s", client.Host, "asset")
	if client.ShowRequests {
		log.Infof("URL: %s\nRequest: %s", url, string(requestBytes))
	}

	requestReader := bytes.NewReader(requestBytes)
	resp, err := http.Post(url, "application/json", requestReader)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 200 {
		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			return "", err
		}
		return tmpFile.Name(), nil
	} else {
		return "", errors.New(fmt.Sprintf("Could not download asset (%d)",
			resp.StatusCode))
	}
}
