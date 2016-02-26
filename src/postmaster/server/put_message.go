package server

import (
	"conduit/log"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"postmaster/api"
	"postmaster/mailbox"
	"time"
)

// putMessage is used to deploy messages to mailboxes, etiher by a list of
// mailboxes or a pattern. Messages are organized into a deployment and
// persisted to the database. If any receiptients are currently polling the
// message will be forwarded to that session.
func putMessage(w http.ResponseWriter, r *http.Request) {
	var request api.PutMessageRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not parse request")
	}
	mailboxes := []mailbox.Mailbox{}
	if request.Pattern != "" {
		results, err := mailbox.Search(request.Pattern)
		if err != nil {
			sendError(w, err.Error())
			return
		}
		for _, mb := range results {
			mailboxes = append(mailboxes, mb)
		}
	}
	for _, mbId := range request.Mailboxes {
		mb, err := mailbox.Find(mbId)
		if err != nil {
			sendError(w, err.Error())
		}
		if mb == nil {
			sendError(w, fmt.Sprintf("Mailbox not found (%s)", mbId))
			return
		}
		mailboxes = append(mailboxes, *mb)
	}

	if len(mailboxes) == 0 {
		sendError(w, "No mailboxes specified")
		return
	}

	if request.Asset != "" {
		assetPath := filepath.Join(filesPath(), request.Asset)
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			sendError(w, "Asset does not exist on server")
			return
		}
	}

	mbList := []string{}
	// dep, err := mailbox.CreateDeployment(request.DeploymentName, request.Token,
	// 	request.Body)

	dep := mailbox.Deployment{
		Name:        request.DeploymentName,
		DeployedBy:  request.AccessKeyName,
		MessageBody: request.Body,
		Asset:       request.Asset,
	}

	err = dep.Create()

	if err != nil {
		sendError(w, err.Error())
		return
	}

	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if err != nil || accessKey == nil {
		sendError(w, "Access key is invalid")
		return
	}

	if !request.Validate(accessKey.Secret) {
		sendError(w, "Signature not valid")
		return
	}

	for _, mb := range mailboxes {
		if !accessKey.CanPut(&mb) {
			sendError(w, "Not allowed to send messages to "+mb.Id)
			return
		}
		var msg *mailbox.Message
		msg, err = dep.Deploy(&mb)
		mbList = append(mbList, mb.Id)
		if err != nil {
			sendError(w, err.Error())
			return
		}
		if pollChannel, ok := pollingChannels[mb.Id]; ok {
			time.Sleep(50 * time.Millisecond)
			pollChannel <- msg
		}
	}

	resp := api.PutMessageResponse{
		MessageSize: r.ContentLength,
		Mailboxes:   mbList,
		Deployment:  dep.Id,
	}
	resp.Sign(accessKey.Name, accessKey.Secret)

	log.Infof("Message received for %d mailboxes from %s", len(mbList),
		dep.DeployedBy)
	writeResponse(&w, resp)
}
