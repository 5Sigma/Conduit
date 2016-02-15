package server

import (
	"net/http"
	"postmaster/api"
	"postmaster/mailbox"
)

// systemStats is json endpoint used to retrieve overall statistics. The token
// used to request the endpoint must have write priviledges.
func systemStats(w http.ResponseWriter, r *http.Request) {
	var request api.SimpleRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not get stats")
		return
	}

	accessKey, _ := mailbox.FindKeyByName(request.AccessKeyName)

	if accessKey == nil {
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

	stats, err := mailbox.Stats()
	if err != nil {
		sendError(w, err.Error())
		return
	}

	response := api.SystemStatsResponse{
		TotalMailboxes:   stats.MailboxCount,
		PendingMessages:  stats.PendingMessages,
		ConnectedClients: int64(len(pollingChannels)),
	}
	response.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, response)
}
