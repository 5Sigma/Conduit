package server

import (
	"conduit/log"
	"fmt"
	"net/http"
	"postmaster/api"
	"postmaster/mailbox"
	"time"
)

// getMessage is used by clients to poll for mailbox messages. This method will
// search the database for messages for the given mailbox. If the mailbox is
// empty and long_polling is enabled it will create a channel and add it to
// pollingChannels. It will then wait for a message to be pushed to that
// channel. It will then continue to output that message. Messages are pushed by
// the putMessage method.
func getMessage(w http.ResponseWriter, r *http.Request) {
	if !EnableLongPolling {
		time.Sleep(ThrottleDelay)
	}
	var request api.GetMessageRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	validateVersion(w, request.Version)

	log.Info("Message request for " + request.Mailbox)

	mb, err := mailbox.Find(request.Mailbox)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	if mb == nil {
		log.Errorf("Could not find a mailbox named '%s'", request.Mailbox)
		sendError(w, fmt.Sprintf("Mailbox %s not found.", request.Mailbox))
		return
	}

	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if accessKey == nil {
		log.Errorf("Could not find an access key named '%s'", request.AccessKeyName)
		sendError(w, "Access key is invalid")
		return
	}

	if !request.Validate(accessKey.Secret) {
		sendError(w, "Signature is invalid")
		return
	}

	if !accessKey.CanGet(mb) {
		sendError(w, "Not allowed to get messages from this mailbox.")
		return
	}

	if err := mb.Checkin(r.RemoteAddr, request.Version); err != nil {
		sendError(w, err.Error())
		return
	}

	msg, err := mb.GetMessage()
	if err != nil {
		sendError(w, err.Error())
		return
	}

	if EnableLongPolling == true && msg == nil {
		pollingChannels[mb.Id] = make(chan *mailbox.Message)
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(100 * time.Second)
			timeout <- true
		}()
		select {
		case m := <-pollingChannels[mb.Id]:
			msg = m
		case <-timeout:
			writeResponse(&w, &api.GetMessageResponse{})
			delete(pollingChannels, mb.Id)
			return
		}
		delete(pollingChannels, mb.Id)
	}

	if msg == nil {
		writeResponse(&w, nil)
		return
	}

	dep, err := msg.GetDeployment()
	if err != nil {
		sendError(w, err.Error())
		return
	}

	response := api.GetMessageResponse{
		Message:      msg.Id,
		Body:         msg.Body,
		CreatedAt:    msg.CreatedAt,
		ReceiveCount: msg.ReceiveCount,
		Deployment:   msg.Deployment,
		Asset:        dep.Asset,
	}

	response.Sign(accessKey.Name, accessKey.Secret)
	log.Infof("Delivering message %s", response.Message)

	writeResponse(&w, response)
}
