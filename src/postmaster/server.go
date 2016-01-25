package postmaster

import (
	"encoding/json"
	"net/http"
	"postmaster/mailbox"
	"regexp"
	"time"
)

var mailboxes []mailbox.Mailbox

type ResponseError struct {
	Error string
}

type EndPoint struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type EndPointHandler struct {
	endpoints []*EndPoint
}

func (h *EndPointHandler) Add(str string,
	handler func(http.ResponseWriter, *http.Request)) {
	rx, _ := regexp.Compile(str)
	h.endpoints = append(h.endpoints, &EndPoint{rx, http.HandlerFunc(handler)})
}

func (h *EndPointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, endpoint := range h.endpoints {
		if endpoint.pattern.MatchString(r.URL.Path) {
			endpoint.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func Start(addr string) {
	// err := mailbox.CreateDB()
	// if err != nil {
	// 	panic(err)
	// }
	endpoints := EndPointHandler{}
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	endpoints.Add(`/mailbox/[0-9a-z\-]+`, getMessage)
	endpoints.Add("/stats", stats)
	http.Handle("/", &endpoints)
	server.ListenAndServe()
}

func getMessage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.String()[9:]

	mb, err := mailbox.Find(id)
	if err != nil {
		e := &ResponseError{Error: err.Error()}
		json.NewEncoder(w).Encode(e)
		return
	}
	if mb == nil {
		e := &ResponseError{Error: "Mailbox not found."}
		json.NewEncoder(w).Encode(e)
		return
	}
	msg, err := mb.GetMessage()
	if err != nil {
		e := &ResponseError{Error: err.Error()}
		json.NewEncoder(w).Encode(e)
	} else {
		json.NewEncoder(w).Encode(msg)
	}
}

func stats(w http.ResponseWriter, r *http.Request) {
	mbxs, err := mailbox.All()
	if err != nil {
		e := &ResponseError{Error: err.Error()}
		json.NewEncoder(w).Encode(e)
	} else {
		json.NewEncoder(w).Encode(mbxs)
	}
}
