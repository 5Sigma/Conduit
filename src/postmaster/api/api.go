package api

import (
	"conduit/info"
	"conduit/log"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/nu7hatch/gouuid"
	"time"
)

type (
	ApiRequest struct {
		Version       string `json:"version"`
		Signature     string `json:"signature"`
		RequestTime   string `json:"requestTime"`
		AccessKeyName string `json:"keyName"`
		Token         string `json:"token"`
	}

	ApiError struct {
		Error string `json:"error"`
	}

	PutMessageRequest struct {
		ApiRequest
		Mailboxes      []string `json:"mailboxes"`
		Body           string   `json:"body"`
		Pattern        string   `json:"pattern"`
		DeploymentName string   `json:"deploymentName"`
		Asset          string   `json:"asset"`
	}

	PutMessageResponse struct {
		ApiRequest
		MessageSize    int64    `json:"messageSize"`
		Mailboxes      []string `json:"mailboxes"`
		Deployment     string   `json:"deploymentId"`
		DeploymentName string   `json:"deploymentName"`
	}

	GetMessageRequest struct {
		ApiRequest
		Mailbox string `json:"mailbox"`
	}

	GetMessageResponse struct {
		ApiRequest
		Message      string    `json:"message"`
		Body         string    `json:"body"`
		CreatedAt    time.Time `json:"createdAt"`
		ReceiveCount int64     `json:"receiveCount"`
		Deployment   string    `json:"deployment"`
		Asset        string    `json:"asset"`
	}
	DeleteMessageRequest struct {
		ApiRequest
		Message string `json:"message"`
	}

	DeleteMessageResponse struct {
		ApiRequest
		Message string `json:"message"`
	}

	SimpleRequest struct {
		ApiRequest
	}

	SimpleResponse struct {
		ApiRequest
		Success bool `json:"success"`
	}

	SystemStatsResponse struct {
		ApiRequest
		TotalMailboxes   int64  `json:"totalMailboxes"`
		MessageCount     int64  `json:"messageCount"`
		PendingMessages  int64  `json:"pendingMessages"`
		ConnectedClients int64  `json:"connectedClients"`
		DBVersion        string `json:"dbVersion"`
		Threads          int64  `json:"threads"`
		CPUCount         int64  `json:"cpuCount"`
		MemoryAllocated  uint64 `json:"memory"`
		Lookups          uint64 `json:lookups"`
		NextGC           uint64 `json:"garbageCollectionAt"`
		FileStoreCount   int64  `json:"filesCount"`
		FileStoreSize    int64  `json:"filesSize"`
	}

	ClientStatusResponse struct {
		ApiRequest
		Clients []ClientStatus
	}

	ClientStatus struct {
		Mailbox  string    `json:"mailbox"`
		LastSeen time.Time `json:"lastSeenAt"`
		Version  string    `json:"version"`
		Host     string    `json:"host"`
		Online   bool      `json:"online"`
	}

	DeploymentStatsRequest struct {
		ApiRequest
		Deployment   string `json:"deploymentId"`
		GetResponses bool   `json:"getResponses"`
		Count        int64  `json:"count"`
		NamePattern  string `json:"nameSearch"`
		TokenPattern string `json:"keyNameSearch"`
	}

	DeploymentStatsResponse struct {
		ApiRequest
		Deployments []DeploymentStats `json:"deployments"`
	}

	DeploymentStats struct {
		ApiRequest
		Id            string               `json:"deploymentId"`
		Name          string               `json:"name"`
		CreatedAt     time.Time            `json:"createdAt"`
		PendingCount  int64                `json:"pendingMessages"`
		MessageCount  int64                `json:"totalMessages"`
		ResponseCount int64                `json:"responseCount"`
		Responses     []DeploymentResponse `json:"repsonses"`
		DeployedBy    string               `json:"deployedBy"`
	}

	DeploymentResponse struct {
		ApiRequest
		Mailbox     string    `json:"mailbox"`
		Response    string    `json:"response"`
		RespondedAt time.Time `json:"respondedAt"`
		IsError     bool      `json:"isError"`
	}

	ResponseRequest struct {
		ApiRequest
		Response string `json:"response"`
		Message  string `json:"mailbox"`
		Error    bool   `json:"Error"`
	}

	RegisterRequest struct {
		ApiRequest
		Mailbox string `json:"mailbox"`
	}

	RegisterResponse struct {
		ApiRequest
		Mailbox         string `json:"mailboxName"`
		AccessKeyName   string `json:"accessKeyName"`
		AccessKeySecret string `json:"accessKeySecret"`
	}

	AgentRequest struct {
		ApiRequest
		Function string
	}

	AgentResponse struct {
		ApiRequest
		Success bool
		Error   string
	}

	UploadFileRequest struct {
		ApiRequest
		Filename string `json:"filename"`
		MD5      string `json:"md5"`
	}
	CheckFileRequest struct {
		ApiRequest
		MD5 string `json:"md5"`
	}
	GetAssetRequest struct {
		ApiRequest
		MD5 string `json:"md5"`
	}
)

func (r *GetMessageResponse) IsEmpty() bool {
	if r.Body == "" {
		return true
	} else {
		return false
	}
}

func (request *ApiRequest) Sign(keyName, secret string) {
	uuid, _ := uuid.NewV4()
	token := uuid.String()
	request.AccessKeyName = keyName
	request.RequestTime = time.Now().Format(time.RFC3339)
	request.Token = token
	request.Version = info.ConduitVersion
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	sig := token + request.RequestTime
	h.Write([]byte(sig))
	request.Signature = base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (request *ApiRequest) Validate(secret string) bool {
	t, err := time.Parse(time.RFC3339, request.RequestTime)
	if err != nil {
		log.Error("Could not parse requestTime")
		return false
	}
	if time.Since(t) > 10*time.Minute {
		log.Error("Time is too far out of sync")
		return false
	}
	key := []byte(secret)
	sig := request.Token + request.RequestTime
	data := []byte(sig)
	signature, err := base64.StdEncoding.DecodeString(request.Signature)
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(signature, expectedMAC)
}
