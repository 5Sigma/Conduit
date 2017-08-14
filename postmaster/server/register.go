package server

import (
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/api"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"net/http"
)

// register is used by administrative clients to reigster new mailboxes.
func register(w http.ResponseWriter, r *http.Request) {
	var request api.RegisterRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if accessKey == nil {
		sendError(w, "Access key is invalid")
		return
	}

	if !request.Validate(accessKey.Secret) {
		sendError(w, "Could not validate signature")
		return
	}

	if !accessKey.CanAdmin() {
		sendError(w, "Not allowed to register mailboxes.")
		return
	}

	if mailbox.KeyExists(request.Mailbox) {
		sendError(w, "An access key already exists with that mailbox name")
		return
	}

	mb, err := mailbox.Create(request.Mailbox)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	mbKey := &mailbox.AccessKey{
		MailboxId: mb.Id,
	}

	err = mbKey.Create()
	if err != nil {
		sendError(w, err.Error())
		return
	}

	resp := api.RegisterResponse{
		Mailbox:         mb.Id,
		AccessKeyName:   mbKey.Name,
		AccessKeySecret: mbKey.Secret,
	}
	resp.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, resp)
	log.Infof("Mailbox %s registered.", mb.Id)
}
