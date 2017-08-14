package server

import (
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/api"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"net/http"
)

// deployRespond is used by scripts to "respond" with information from the
// remote system. These responses can then be reported to the admin client that
// deployed the original script.
func deployRespond(w http.ResponseWriter, r *http.Request) {
	var request api.ResponseRequest
	err := readRequest(r, &request)

	if err != nil {
		sendError(w, "Could not parse request")
		return
	}
	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if accessKey == nil {
		sendError(w, "Access key is not valid")
		return
	}

	if !request.Validate(accessKey.Secret) {
		sendError(w, "Could not validate signature")
		return
	}

	msg, err := mailbox.FindMessage(request.Message)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	if msg == nil {
		sendError(w, "Could not find message "+request.Message)
		return
	}

	mb, err := mailbox.Find(msg.Mailbox)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	if mb == nil {
		sendError(w, "Mailbox not found")
		return
	}

	dep, err := mailbox.FindDeployment(msg.Deployment)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	if dep == nil {
		sendError(w, "Could not find deployment "+msg.Deployment)
		return
	}

	if !accessKey.CanGet(mb) {
		sendError(w, "Not allowed to respond to deploys")
		return
	}
	err = dep.AddResponse(msg.Mailbox, request.Response, request.Error)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	response := api.SimpleResponse{Success: true}
	response.Sign(accessKey.Name, accessKey.Secret)
	log.Infof("Reponse added to %s from %s", dep.Id, msg.Mailbox)
	writeResponse(&w, response)
}
