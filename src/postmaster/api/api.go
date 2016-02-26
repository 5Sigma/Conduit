// This package contains structures used for json serialization/deserialization
// of server/client requests and responses. Each request inherits ApiRequest
// which provides the basic fields on all requests and responses. It also
// provides functionality to sign and validate them using HMAC signatures.
package api

import (
	"conduit/info"
	"conduit/log"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/nu7hatch/gouuid"
	"sort"
	"time"
)

type (
	// ApiRequest is used by all requests and responses to hold global fields. It
	// also handles request/response signing.
	ApiRequest struct {
		Version       string `json:"version"`
		Signature     string `json:"signature"`
		RequestTime   string `json:"requestTime"`
		AccessKeyName string `json:"keyName"`
		Token         string `json:"token"`
	}

	// ApiError is returned by the server in place of a response whenever an error
	// occures.
	ApiError struct {
		Error string `json:"error"`
	}

	// PutMessageRequest is used by a client to deploy a message to a mailbox or
	// list of mailboxes.
	PutMessageRequest struct {
		ApiRequest
		// Mailboxes is a slice of specific mailboxes to delier the message to. This
		// value is not exclusive with Pattern.
		Mailboxes []string `json:"mailboxes"`
		// Body holds the content of the message.
		Body string `json:"body"`
		// Pattern is a glob enabled pattern of messages to deilver to. '*' is used
		// as a wildcard. Ex: cluster.servers.*
		Pattern string `json:"pattern"`
		// DeploymentName is an optional name for the deployment. If this value is
		// empty a randomized identifier will be used.
		DeploymentName string `json:"deploymentName"`
		// Asset is an MD5 hash of an asset on the server that should be delivered
		// with the message.
		Asset string `json:"asset"`
	}

	// PutMessageResponse is used when the server responds to a PutMessagRequest.
	PutMessageResponse struct {
		ApiRequest
		// The size of the message in bytes. This is only the size of the message
		// body, not including any assets.
		MessageSize int64 `json:"messageSize"`
		// Mailboxes is the a string slice of all mailboxes that were deployed to.
		// This includes any mailboxes provided specifically and any mailboxes
		// matched through a given pattern.
		Mailboxes []string `json:"mailboxes"`
		// Deployment is the internal identifier for the deployment. This fields is
		// also copied to the DeploymentName if one was not given.
		Deployment string `json:"deploymentId"`
		// The custom name for the deployment if one was provided with the request
		DeploymentName string `json:"deploymentName"`
	}

	// GetMessageRequest is used when the client polls for a message.
	GetMessageRequest struct {
		ApiRequest
		// Mailbox is the mailbox the client is polling for. the AccessKey used
		// must be one that is linked to the mailbox specified.
		Mailbox string `json:"mailbox"`
	}

	// GetMessageResponse is sent when the server reponds to a message request.
	GetMessageResponse struct {
		ApiRequest
		// Message holds the message identifier for the message. This identifier can
		// be used by the client to mark the message as deleted.
		Message string `json:"message"`
		// Body contains the content of the message.
		Body string `json:"body"`
		// CreatedAt is the date and time when the PutMessage request was made.
		CreatedAt time.Time `json:"createdAt"`
		// ReceiveCount is the number of times this message has been received. Each
		// time a GetMessageResponse is sent to the client this value is
		// incrememnted by one.
		ReceiveCount int64 `json:"receiveCount"`
		// Deployment is the identifier for the deploymennt this message belongs to.
		Deployment string `json:"deployment"`
		// Asset has the MD5 value of a file linked to this message. If the value is
		// present the client can use this to value to download the file before
		// processing the message.
		Asset string `json:"asset"`
	}

	// DeleteMessageRequest is used by the client to mark a message as deleted.
	DeleteMessageRequest struct {
		ApiRequest
		// The message identifier given with the GetMessageResponse for the message.
		Message string `json:"message"`
	}

	// DeleteMessageResponse is used when the server confirms a message's
	// deletion.
	DeleteMessageResponse struct {
		ApiRequest
		// message holds the message identifier.
		Message string `json:"message"`
	}

	// SimpleRequest is a catch all request used when no special parameters are
	// required.
	SimpleRequest struct {
		ApiRequest
	}

	// SimpleResponse is a catch all response used when no additional information
	// is needed.
	SimpleResponse struct {
		ApiRequest
		// Success is a general success notification.
		Success bool `json:"success"`
	}

	// SystemStatsResponse is used to provide clients with various system level
	// metrics and information.
	SystemStatsResponse struct {
		ApiRequest
		// TotalMailboxes is a count of all mailboxes currently registered with the
		// system.
		TotalMailboxes int64 `json:"totalMailboxes"`
		// MessageCount is a count of all messages, deleted or not, that have passed
		// through the system.
		MessageCount int64 `json:"messageCount"`
		// PendingMessages is a count of all messages that have not been marked as
		// deleted.
		PendingMessages int64 `json:"pendingMessages"`
		// ConnectedClients is a count of all clients that are currently connected
		// to the server. Due to the nature of the long polling this value is
		// accurate within 5-9 minutes.
		ConnectedClients int64 `json:"connectedClients"`
		// DBVersion is the internal version of the database. When a new version of
		// the server is used it may upgrade the database information to be
		// compatible.
		DBVersion string `json:"dbVersion"`
		// Threads is a count of open goroutines the system is currently using.
		Threads int64 `json:"threads"`
		// CPUCount is the number of CPUs available to the server.
		CPUCount int64 `json:"cpuCount"`
		// MemoryAllocated is the amount of memory in bytes that is in use by the
		// server. Not including memory that has been marked freed.
		MemoryAllocated uint64 `json:"memory"`
		// Lookups is the total number of memory lookups since the server was
		// started.
		Lookups uint64 `json:lookups"`
		// NextGC is an amount of memory, in bytes. When this amount is allocated by
		// the server garbage collection will be run.
		NextGC uint64 `json:"garbageCollectionAt"`
		// FileStoreCount is the number of message assets currently stored on the
		// server. These assets are removed once an hour if they are no more pending
		// messages attached to them.
		FileStoreCount int64 `json:"filesCount"`
		// FileStoreSize is the total size in bytes of all message assets currently
		// on the server.
		FileStoreSize int64 `json:"filesSize"`
	}

	// ClientStatusResponse sends back information about mailboxes, providing
	// their version and connected status.
	ClientStatusResponse struct {
		ApiRequest
		// The list of clients in the system and their information.
		Clients ClientStatusCollection
	}

	// ClientStatus holds information about a given client/mailbox
	ClientStatus struct {
		// Mailbox is the name of the mailbox
		Mailbox string `json:"mailbox"`
		// LastSeen is the date and time a client last checked this mailbox for a
		// message.
		LastSeen time.Time `json:"lastSeenAt"`
		// Version is the conduit version reported the last time a client checked
		// this mailbox for a message.
		Version string `json:"version"`
		// Host is the remote IP Address of the last client to check this mailbox
		// for a message.
		Host string `json:"host"`
		// Online is true if the client is currently connected and waiting for a
		// message.
		Online bool `json:"online"`
	}

	//ClientStatusCollection is a slice of type ClientStatus that implements
	//sort.Interface. This structure is used within the ClientStatusResponse to
	//list information about clients in the system.
	ClientStatusCollection []ClientStatus

	// DeploymentStatsRequest is used to get information about a given deployment.
	DeploymentStatsRequest struct {
		ApiRequest
		// Deployment is the identifier for the deployment. This value must be
		// provided unless NamePattern is provided.
		Deployment string `json:"deploymentId"`
		// GetResopnses controls if responses are returned along with the deployment
		// information.
		GetResponses bool `json:"getResponses"`
		// Count is the number of deployments to return in the response.
		Count int64 `json:"count"`
		// NamePattern is a wildcard search for deployment names.
		NamePattern string `json:"nameSearch"`
		// TokenPattern is a wildcard search for deployments by AccessKeyName
		TokenPattern string `json:"keyNameSearch"`
	}

	// DeploymentStatsResponse is used to respond to a DeploymentStatsRequest with
	// information about a number of deployments.
	DeploymentStatsResponse struct {
		ApiRequest
		// Deployments is an array of the deployments that match the parameters
		// given by the DeploymentStatsRequest
		Deployments []DeploymentStats `json:"deployments"`
	}

	// DeploymentStats holds information about a single deployment. It is used by
	// the DeploymentStatsResponse.
	DeploymentStats struct {
		ApiRequest
		// Id is the identifier for a given deployment.
		Id string `json:"deploymentId"`
		// Name is the name for the deployment if one was provided when it was
		// created.
		Name string `json:"name"`
		// CreatedAt is the date and time the deployment was created.
		CreatedAt time.Time `json:"createdAt"`
		// PendingCount is the number of messages this deployment has that have not
		// been marked deleted.
		PendingCount int64 `json:"pendingMessages"`
		// MessageCount is the total number of messages linked to this deployment.
		MessageCount int64 `json:"totalMessages"`
		// ResponseCount is the number of responses that have been recoreded for
		// this deployment.
		ResponseCount int64 `json:"responseCount"`
		// Responses is a collection of all responses from for the deployment. This
		// value is only provided if the request marked GetResponses as true and a
		// deployment was requested by Id.
		Responses DeploymentResponses `json:"repsonses"`
		// DeployedBy is the name of the AccessKey used to create the deployment.
		DeployedBy string `json:"deployedBy"`
	}

	// DeploymentResponse is used as part of the DeploymentStats and contains
	// information about an individual response to a deployment.
	DeploymentResponse struct {
		ApiRequest
		// Mailbox is the id of the mailbox that responded.
		Mailbox string `json:"mailbox"`
		// Response is the content of the response.
		Response string `json:"response"`
		// RespondedAt is the date and time the response was received.
		RespondedAt time.Time `json:"respondedAt"`
		// IsError is a value provided with the response indicating if this is an
		// error response.
		IsError bool `json:"isError"`
	}

	// DeploymentResponses is a collection of type DeplyomentResponse that
	// implements sort.Interface.
	DeploymentResponses []DeploymentResponse

	// ResponseRequest is used by the client to provie a response to a message.
	ResponseRequest struct {
		ApiRequest
		// Response is the content of the response.
		Response string `json:"response"`
		// Message is the message identifier the client is responding to
		Message string `json:"mailbox"`
		// Error indicates if the response is an error or not.
		Error bool `json:"Error"`
	}

	// RegisterRequest is used by adminstrative clients to register a new mailbox.
	RegisterRequest struct {
		ApiRequest
		// The name of the mailbox to register.
		Mailbox string `json:"mailbox"`
	}

	// RegisterResponse is used by the server to respond to a RegisterRequest
	RegisterResponse struct {
		ApiRequest
		// Mailbox is the name of the mailbox that was registered.
		Mailbox string `json:"mailboxName"`
		// AccessKeyName is the name of the access key generated for the mailbox.
		// This is always the mailboxe's name.
		AccessKeyName string `json:"accessKeyName"`
		// The AccessKeySecret generated for the mailbox. This is used in the config
		// file for clients accessing this mailbox. It is used to cryptographically
		// sign and validate messages.
		AccessKeySecret string `json:"accessKeySecret"`
	}

	// AgentRequest is used by instances running in agent mode. This request
	// causes the agent to process the given function.
	AgentRequest struct {
		ApiRequest
		// Function is a string containing the function literal to process.
		Function string
	}

	// AgentResponse is used by agents to respond once an AgentRequest has been
	// received and processed.
	AgentResponse struct {
		ApiRequest
		// Success is true if the passed function does not throw any errors.
		Success bool `json:"success"`
		// Error contains the exception message of any thrown errors.
		Error string `json:"error"`
	}

	// UploadFileRequest is used by clients to upload a message asset.
	UploadFileRequest struct {
		ApiRequest
		// Filename is the name of the file being uploaded.
		Filename string `json:"filename"`
		// MD5 is a hex string of the md5 hash for the asset file.
		MD5 string `json:"md5"`
	}

	// CheckFileRequest is used to check if a file exists on the server before
	// uploading. This is used to prevent uploading the same file multiple times.
	CheckFileRequest struct {
		ApiRequest
		// MD5 is a hex string of the md5 hash for the asset file.
		MD5 string `json:"md5"`
	}

	// GetAssetRequest is used by cleints to download an asset.
	GetAssetRequest struct {
		ApiRequest
		// MD5 is a hex string of the md5 hash for the asset file. This hash is
		// provedied with the DeploymentResponse when an asset is attached to a
		// deployment.
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

// Sign is used to sign a request or a response with a given access key. It uses
// a randomly generated string combined with the current time. This value is
// then used to generated a HMAC signature, which is attached to the structure.
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

// Validate uses the token and request time of the request to calculate a HMAC
// signature for the request. This is then compared to the one provided with the
// message to validate its authenticity.
func (request *ApiRequest) Validate(secret string) bool {
	t, err := time.Parse(time.RFC3339, request.RequestTime)
	if err != nil {
		log.Debug("Could not parse requestTime")
		return false
	}
	if time.Since(t) > 30*time.Minute {
		log.Debug("Time is too far out of sync")
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

// Implmentation for sort.Interface
func (responses DeploymentResponses) Len() int { return len(responses) }
func (responses DeploymentResponses) Swap(i, j int) {
	responses[i], responses[j] = responses[j], responses[i]
}
func (responses DeploymentResponses) Less(i, j int) bool {
	return responses[i].Mailbox < responses[j].Mailbox
}
func (responses DeploymentResponses) Sort() {
	sort.Sort(responses)
}

// Implementation for sort.Interface
func (col ClientStatusCollection) Len() int { return len(col) }
func (col ClientStatusCollection) Swap(i, j int) {
	col[i], col[j] = col[j], col[i]
}
func (col ClientStatusCollection) Less(i, j int) bool {
	return col[i].Mailbox < col[j].Mailbox
}
func (col ClientStatusCollection) Sort() {
	sort.Sort(col)
}
