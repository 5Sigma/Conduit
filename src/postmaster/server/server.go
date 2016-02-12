package server

import (
	"conduit/log"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"postmaster/api"
	"postmaster/mailbox"
	"regexp"
	"time"
)

// pollingChannels is used to internally store a map of connected clients that
// are using long polling. A message is sent to the channel to complete the poll
// cycle.
var pollingChannels = make(map[string]chan *mailbox.Message)

// EnableLongPolling controls if long polling is used. If false than clients
// will be given an immediate empty repsonse if no messages are waiting for
// them. If true the connection will be held until a message comes in or until
// it timesout.
var EnableLongPolling = true

// ThrottleDelay will delay messages from being pushed to clients. This will
// artificially throttle the connect and reconnects from the clients.
var ThrottleDelay = 500 * time.Millisecond

var serverRunning = false

type EndPoint struct {
	Method  string
	Regex   *regexp.Regexp
	Handler http.Handler
}

type EndPointHandler struct {
	endpoints []*EndPoint
}

func (h *EndPointHandler) Add(method string, pattern string,
	handler func(http.ResponseWriter, *http.Request)) {
	rx, _ := regexp.Compile(pattern)
	ep := &EndPoint{
		Method:  method,
		Regex:   rx,
		Handler: http.HandlerFunc(handler),
	}
	h.endpoints = append(h.endpoints, ep)
}

func (h *EndPointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, endpoint := range h.endpoints {
		if endpoint.Regex.MatchString(r.URL.Path) && endpoint.Method == r.Method {
			endpoint.Handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func Start(addr string) error {
	if serverRunning == true {
		return nil
	}

	if mailbox.DB == nil {
		mailbox.OpenDB()
		if _, err := os.Stat("mailboxes.db"); os.IsNotExist(err) {
			err := mailbox.CreateDB()
			if err != nil {
				panic(err)
			}
		}
	}
	endpoints := EndPointHandler{}
	svr := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
	}
	endpoints.Add("POST", `/get`, getMessage)
	endpoints.Add("POST", "/put", putMessage)
	endpoints.Add("POST", "/stats/clients", clientStats)
	endpoints.Add("POST", "/stats", systemStats)
	endpoints.Add("POST", "/delete", deleteMessage)
	endpoints.Add("POST", "/deploy/list", deployInfo)
	endpoints.Add("POST", "/deploy/respond", deployRespond)
	endpoints.Add("POST", "/register", register)
	endpoints.Add("POST", "/deregister", deregister)
	http.Handle("/", &endpoints)
	serverRunning = true
	err := svr.ListenAndServe()
	return err
}

func CreateMailbox(id string) (*mailbox.Mailbox, error) {
	return mailbox.Create(id)
}

func writeResponse(w *http.ResponseWriter, response interface{}) error {
	bytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	io.WriteString(*w, string(bytes))
	return nil
}

func readRequest(r *http.Request, req interface{}) error {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, req)
	return err
}

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

	msg, err := mb.GetMessage()
	if err != nil {
		e := &api.ApiError{Error: err.Error()}
		writeResponse(&w, e)
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

	response := api.GetMessageResponse{
		Message:      msg.Id,
		Body:         msg.Body,
		CreatedAt:    msg.CreatedAt,
		ReceiveCount: msg.ReceiveCount,
		Deployment:   msg.Deployment,
	}
	response.Sign(accessKey.Name, accessKey.Secret)
	log.Infof("Delivering message %s", response.Message)

	writeResponse(&w, response)
}

func sendError(w http.ResponseWriter, msg string) {
	e := &api.ApiError{
		Error: msg,
	}
	w.WriteHeader(http.StatusBadRequest)
	writeResponse(&w, e)
}

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

	mbList := []string{}
	// dep, err := mailbox.CreateDeployment(request.DeploymentName, request.Token,
	// 	request.Body)

	dep := mailbox.Deployment{
		Name:        request.DeploymentName,
		DeployedBy:  request.AccessKeyName,
		MessageBody: request.Body,
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

	clients := make(map[string]bool)
	mbxs, err := mailbox.All()
	if err != nil {
		sendError(w, err.Error())
		return
	}
	for _, mb := range mbxs {
		if _, ok := pollingChannels[mb.Id]; ok {
			clients[mb.Id] = true
		} else {
			clients[mb.Id] = false
		}
	}
	response := api.ClientStatusResponse{Clients: clients}
	response.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, response)
}

func deployInfo(w http.ResponseWriter, r *http.Request) {
	var request api.DeploymentStatsRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not parse request")
		return
	}

	if request.NamePattern == "" {
		request.NamePattern = ".*"
	}

	if request.TokenPattern == "" {
		request.TokenPattern = ".*"
	}

	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if err != nil || accessKey == nil {
		sendError(w, "Access key is invalid")
		return
	}
	if !accessKey.CanAdmin() {
		sendError(w, "Not allowed to list deployments")
		return
	}
	if !request.Validate(accessKey.Secret) {
		sendError(w, "Signature invalid")
		return
	}
	response := api.DeploymentStatsResponse{}
	if request.Deployment == "" {
		log.Infof("Listing all deploys for %s", accessKey.Name)
		deployments, err := mailbox.ListDeployments(request.NamePattern,
			int(request.Count), request.TokenPattern)
		if err != nil {
			sendError(w, err.Error())
			return
		}
		for _, d := range deployments {
			dStats, err := d.Stats()
			if err != nil {
				sendError(w, err.Error())
				return
			}
			statsResp := api.DeploymentStats{
				Name:          d.Name,
				Id:            d.Id,
				PendingCount:  dStats.PendingCount,
				MessageCount:  dStats.MessageCount,
				ResponseCount: dStats.ResponseCount,
				CreatedAt:     d.DeployedAt,
				DeployedBy:    d.DeployedBy,
			}
			response.Deployments = append(response.Deployments, statsResp)
		}
	} else {
		dep, err := mailbox.FindDeployment(request.Deployment)
		if err != nil {
			sendError(w, err.Error())
			return
		}

		if dep == nil {
			sendError(w, "Deployment not found")
			return
		}

		dStats, err := dep.Stats()
		if err != nil {
			sendError(w, err.Error())
			return
		}
		deploymentResponses, err := dep.GetResponses()
		if err != nil {
			sendError(w, err.Error())
			return
		}
		statsResp := api.DeploymentStats{
			Name:          dep.Name,
			Id:            dep.Id,
			PendingCount:  dStats.PendingCount,
			MessageCount:  dStats.MessageCount,
			ResponseCount: dStats.ResponseCount,
			CreatedAt:     dep.DeployedAt,
			Responses:     []api.DeploymentResponse{},
		}
		for _, r := range deploymentResponses {
			apiR := api.DeploymentResponse{
				Mailbox:     r.Mailbox,
				Response:    r.Response,
				RespondedAt: r.RespondedAt,
				IsError:     r.IsError,
			}
			statsResp.Responses = append(statsResp.Responses, apiR)
		}
		response.Deployments = append(response.Deployments, statsResp)
	}
	response.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, response)
}

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
