package server

import (
	"net/http"
	"postmaster/api"
	"postmaster/mailbox"
)

// deregister is used by administrative clients to remove mailboxes. It deletes
// the mailbox and all messages from the database.
func deregister(w http.ResponseWriter, r *http.Request) {
	var request api.RegisterRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	accessKey, _ := mailbox.FindKeyByName(request.AccessKeyName)
	if accessKey == nil {
		sendError(w, "Access key is invalid")
		return
	}
	if !request.Validate(accessKey.Secret) {
		sendError(w, "Could not validate signature")
		return
	}
	if !accessKey.CanAdmin() {
		sendError(w, "Not allowed to deregister mailboxes")
		return
	}
	err = mailbox.Deregister(request.Mailbox)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	resp := api.SimpleResponse{Success: true}
	resp.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, resp)
}
