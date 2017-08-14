package server

import (
	"fmt"
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/api"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"math/rand"
	"net/http"
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

	mb, err := mailbox.Find(request.Mailbox)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	if mb == nil {
		log.Warnf("Could not find a mailbox named '%s'", request.Mailbox)
		sendError(w, fmt.Sprintf("Mailbox %s not found.", request.Mailbox))
		return
	}

	log.Debugf("Message request from %s", mb.Id)

	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if accessKey == nil {
		log.Warnf("Could not find an access key named '%s'", request.AccessKeyName)
		sendError(w, "Access key is invalid")
		return
	}

	if !request.Validate(accessKey.Secret) {
		log.Warnf(fmt.Sprintf("Signature for %s invalid", mb.Id))
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
		log.Errorf("Error retrieving messages: %s", err.Error())
		return
	}

	if EnableLongPolling == true && msg == nil {
		if _, ok := pollingChannels[mb.Id]; ok {
			delete(pollingChannels, mb.Id)
		}
		// Create a channel for the client. This channel has a message pushed to it
		// from the putMessage function. When a message gets delivered.
		pollingChannels[mb.Id] = make(chan *mailbox.Message)
		timeout := make(chan bool, 1)
		// This goroutine will create a timeout to close the long polling connection
		// and force the client to reconnect.
		go func() {
			// Randomize the polling timeout in order to stagger client reconnects.
			sleepTime := rand.Intn(500) + 200
			time.Sleep(time.Duration(sleepTime) * time.Second)
			timeout <- true
		}()
		// Wait for either a timeout or a message to be sent to a channel.
		select {
		case m := <-pollingChannels[mb.Id]:
			msg = m
		case <-timeout:
			response := api.GetMessageResponse{}
			response.Sign(accessKey.Name, accessKey.Secret)
			writeResponse(&w, response)
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
	log.Infof("Delivering message %s to %s", response.Message, mb.Id)

	writeResponse(&w, response)
}
