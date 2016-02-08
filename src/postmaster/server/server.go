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
	http.Handle("/", &endpoints)
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

	mb, err := mailbox.Find(request.Mailbox)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	if mb == nil {
		sendError(w, fmt.Sprintf("Mailbox %s not found.", request.Mailbox))
		return
	}

	if !mailbox.TokenCanGet(request.Token, mb.Id) {
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
		DeployedBy:  request.Token,
		MessageBody: request.Body,
	}

	err = dep.Create()

	if err != nil {
		sendError(w, err.Error())
		return
	}

	for _, mb := range mailboxes {
		if !mailbox.TokenCanPut(request.Token, mb.Id) {
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

	log.Infof("Message received for %d mailboxes", len(mbList))
	writeResponse(&w, resp)
}

func deleteMessage(w http.ResponseWriter, r *http.Request) {
	var request api.DeleteMessageRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not parse json request")
		return
	}
	err = mailbox.DeleteMessage(request.Message)
	if err != nil {
		sendError(w, "Could not delete message")
		return
	}
	resp := api.DeleteMessageResponse{Message: request.Message}
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

	if !mailbox.TokenCanAdmin(request.Token) {
		sendError(w, "Not allowed to get statistics")
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

	writeResponse(&w, response)
}

func clientStats(w http.ResponseWriter, r *http.Request) {
	var request api.SimpleRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Bad request")
	}
	if !mailbox.TokenCanAdmin(request.Token) {
		sendError(w, "Not allowed to get statistics")
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

	if !mailbox.TokenCanAdmin(request.Token) {
		sendError(w, "Not allowed to list deployments")
		return
	}
	response := api.DeploymentStatsResponse{}
	if request.Deployment == "" {
		log.Info("Listing all deploys")
		deployments, err := mailbox.ListDeployments(request.NamePattern,
			int(request.Count), request.TokenPattern)
		if err != nil {
			sendError(w, err.Error())
			return
		}
		fmt.Println(request.TokenPattern)
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
			}
			statsResp.Responses = append(statsResp.Responses, apiR)
		}
		response.Deployments = append(response.Deployments, statsResp)
	}
	writeResponse(&w, response)
}

func deployRespond(w http.ResponseWriter, r *http.Request) {
	var request api.ResponseRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not parse request")
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
	dep, err := mailbox.FindDeployment(msg.Deployment)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	if dep == nil {
		sendError(w, "Could not find deployment "+msg.Deployment)
		return
	}
	if !mailbox.TokenCanGet(request.Token, msg.Mailbox) {
		sendError(w, "Not allowed to respond to deploys")
		return
	}
	err = dep.AddResponse(msg.Mailbox, request.Response)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	response := api.SimpleResponse{Success: true}
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
	if !mailbox.TokenCanAdmin(request.Token) {
		sendError(w, "Not allowed to register mailboxes.")
		return
	}
	mb, err := mailbox.Create(request.Mailbox)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	token, err := mb.CreateToken()
	if err != nil {
		sendError(w, err.Error())
		return
	}
	resp := api.RegisterResponse{
		Mailbox:      mb.Id,
		MailboxToken: token.Token,
	}
	writeResponse(&w, resp)
	log.Infof("Mailbox %s registered.", mb.Id)
}
