package server

import (
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

var pollingChannels = make(map[string]chan *mailbox.Message)

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
	endpoints.Add("POST", "/delete", deleteMessage)
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
	fmt.Println("REQUESTING MESSAGE")
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

	msg, err := mb.GetMessage()
	if err != nil {
		e := &api.ApiError{Error: err.Error()}
		writeResponse(&w, e)
	}
	if msg == nil {
		pollingChannels[mb.Id] = make(chan *mailbox.Message)
		msg = <-pollingChannels[mb.Id]
	}
	delete(pollingChannels, mb.Id)

	response := api.GetMessageResponse{
		Message:      msg.Id,
		Body:         msg.Body,
		CreatedAt:    msg.CreatedAt,
		ReceiveCount: msg.ReceiveCount,
	}

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
			sendError(w, fmt.Sprintf("Mailbox not found. (%s)", mbId))
			return
		}
		mailboxes = append(mailboxes, *mb)
	}

	if len(mailboxes) == 0 {
		sendError(w, "No mailboxes specified")
		return
	}

	mbList := []string{}
	for _, mb := range mailboxes {
		msg, err := mb.PutMessage(request.Body)
		mbList = append(mbList, mb.Id)
		if err != nil {
			sendError(w, err.Error())
		}
		if pollChannel, ok := pollingChannels[mb.Id]; ok {
			pollChannel <- msg
		}
	}
	resp := api.PutMessageResponse{
		MessageSize: r.ContentLength,
		Mailboxes:   mbList,
	}
	writeResponse(&w, resp)
}

func deleteMessage(w http.ResponseWriter, r *http.Request) {
	var request api.DeleteMessageRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not parse json request.")
		return
	}
	err = mailbox.DeleteMessage(request.Message)
	if err != nil {
		sendError(w, "Could not delete message")
		return
	}
	resp := api.DeleteMessageResponse{Message: request.Message}
	writeResponse(&w, resp)
}

// func mailboxStats(w http.ResponseWriter, r *http.Request) {
// 	id := r.URL.String()[9:45]
// 	mb, err := mailbox.Find(id)
// 	if err != nil {
// 		e := &ApiError{Error: err.Error()}
// 		json.NewEncoder(w).Encode(e)
// 		return
// 	}
// 	if mb == nil {
// 		fmt.Println("id")
// 		e := &ApiError{Error: "Mailbox not found."}
// 		json.NewEncoder(w).Encode(e)
// 		return
// 	}
// 	stats, err := mb.Stats()
// 	if err != nil {
// 		e := &ApiError{Error: err.Error()}
// 		json.NewEncoder(w).Encode(e)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(stats)
// }
//
// func stats(w http.ResponseWriter, r *http.Request) {
// 	mbxs, err := mailbox.All()
// 	if err != nil {
// 		e := &ApiError{Error: err.Error()}
// 		json.NewEncoder(w).Encode(e)
// 	} else {
// 		json.NewEncoder(w).Encode(mbxs)
// 	}
// }
