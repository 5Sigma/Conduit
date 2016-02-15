package server

import (
	"conduit/log"
	"net/http"
	"postmaster/api"
	"postmaster/mailbox"
)

// deleteMessage is used by clients to mark messages as processed. The message
// remains in the database, but is marked deleted and will no longer be
// presented to polling clients.
func deleteMessage(w http.ResponseWriter, r *http.Request) {
	var request api.DeleteMessageRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not parse json request")
		return
	}

	accessKey, _ := mailbox.FindKeyByName(request.AccessKeyName)
	if accessKey == nil {
		sendError(w, "Access key invalid")
		return
	}

	if !request.Validate(accessKey.Secret) {
		sendError(w, "Signature is invalid")
		return
	}

	err = mailbox.DeleteMessage(request.Message)
	if err != nil {
		sendError(w, "Could not delete message")
		return
	}
	resp := api.DeleteMessageResponse{Message: request.Message}
	resp.Sign(accessKey.Name, accessKey.Secret)
	log.Infof("Message %s deleted", request.Message)
	writeResponse(&w, resp)
}
