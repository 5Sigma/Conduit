package api

import (
	"time"
)

type ApiError struct {
	Error string `json:"error"`
}

type PutMessageRequest struct {
	Mailboxes      []string `json:"mailboxes"`
	Body           string   `json:"body"`
	Pattern        string   `json:"pattern"`
	Token          string   `json:"token"`
	DeploymentName string   `json:"deploymentName"`
}

type PutMessageResponse struct {
	MessageSize    int64    `json:"messageSize"`
	Mailboxes      []string `json:"mailboxes"`
	Deployment     string   `json:"deploymentId"`
	DeploymentName string   `json:"deploymentName"`
}

type GetMessageRequest struct {
	Token   string `json:"token"`
	Mailbox string `json:"mailbox"`
}

type GetMessageResponse struct {
	Message      string    `json:"message"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"createdAt"`
	ReceiveCount int64     `json:"receiveCount"`
	Deployment   string    `json:"deployment"`
}

func (r *GetMessageResponse) IsEmpty() bool {
	if r.Body == "" {
		return true
	} else {
		return false
	}
}

type DeleteMessageRequest struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type DeleteMessageResponse struct {
	Message string `json:"message"`
}

type SimpleRequest struct {
	Token string `json:"token"`
}

type SimpleResponse struct {
	Success bool `json:"success"`
}

type SystemStatsResponse struct {
	TotalMailboxes   int64 `json:"totalMailboxes"`
	PendingMessages  int64 `json:"pendingMessages"`
	ConnectedClients int64 `json:"connectedClients"`
}

type ClientStatusResponse struct {
	Clients map[string]bool `json:"clients"`
}

type DeploymentStatsRequest struct {
	Token        string `json:"token"`
	Deployment   string `json:"deploymentId"`
	GetResponses bool   `json:"getResponses"`
	Count        int64  `json:"count"`
	NamePattern  string `json:"nameSearch"`
	TokenPattern string `json:"tokenSearch"`
}

type DeploymentStatsResponse struct {
	Deployments []DeploymentStats `json:"deployments"`
}

type DeploymentStats struct {
	Id            string               `json:"deploymentId"`
	Name          string               `json:"name"`
	CreatedAt     time.Time            `json:"createdAt"`
	PendingCount  int64                `json:"pendingMessages"`
	MessageCount  int64                `json:"totalMessages"`
	ResponseCount int64                `json:"responseCount"`
	Responses     []DeploymentResponse `json:"repsonses"`
	DeployedBy    string               `json:"deployedBy"`
}

type DeploymentResponse struct {
	Mailbox     string    `json:"mailbox"`
	Response    string    `json:"response"`
	RespondedAt time.Time `json:"respondedAt"`
}

type ResponseRequest struct {
	Token    string `json:"token"`
	Response string `json:"response"`
	Message  string `json:"mailbox"`
}

type RegisterRequest struct {
	Mailbox string `json:"mailbox"`
	Token   string `json:"token"`
}

type RegisterResponse struct {
	Mailbox      string `json:"mailboxName"`
	MailboxToken string `json:"accessKey"`
}
