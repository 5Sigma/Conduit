package server

import (
	"net/http"
	"postmaster/api"
	"postmaster/mailbox"
)

// clientStats reports all mailboxes and their current connection status.
func clientStats(w http.ResponseWriter, r *http.Request) {
	var request api.SimpleRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Bad request")
	}

	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if err != nil || accessKey == nil {
		sendError(w, "Access key is invalid")
		return
	}

	if !accessKey.CanAdmin() {
		sendError(w, "Not allowed to get statistics")
		return
	}

	if !request.Validate(accessKey.Secret) {
		sendError(w, "Could not validate signature")
		return
	}

	clients := []api.ClientStatus{}
	mbxs, err := mailbox.All()
	if err != nil {
		sendError(w, err.Error())
		return
	}
	for _, mb := range mbxs {
		st := api.ClientStatus{
			Mailbox:  mb.Id,
			Version:  mb.Version,
			Host:     mb.Host,
			LastSeen: mb.LastSeen,
		}
		if _, ok := pollingChannels[mb.Id]; ok {
			st.Online = true
		} else {
			st.Online = false
		}
		clients = append(clients, st)
	}
	response := api.ClientStatusResponse{Clients: clients}
	response.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, response)
}
