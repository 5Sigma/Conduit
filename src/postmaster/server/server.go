package server

import (
	"conduit/log"
	"encoding/json"
	"github.com/kardianos/osext"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"postmaster/api"
	"postmaster/mailbox"
	"regexp"
	"time"
)

var (
	// pollingChannels is used to internally store a map of connected clients that
	// are using long polling. A message is sent to the channel to complete the poll
	// cycle.
	pollingChannels = make(map[string]chan *mailbox.Message)

	// EnableLongPolling controls if long polling is used. If false than clients
	// will be given an immediate empty repsonse if no messages are waiting for
	// them. If true the connection will be held until a message comes in or until
	// it timesout.
	EnableLongPolling = true

	// ThrottleDelay will delay messages from being pushed to clients. This will
	// artificially throttle the connect and reconnects from the clients.
	ThrottleDelay = 500 * time.Millisecond

	// serverRunning is set to true before the server begins listening. This is
	// used to help testing.
	serverRunning = false
)

type (
	// EndPoint repesents an API endpoint for the server and is used in request
	// routing.
	EndPoint struct {
		Method  string
		Regex   *regexp.Regexp
		Handler http.Handler
	}

	// EndPointHandler is a collection of API endpoints.
	EndPointHandler struct {
		endpoints []*EndPoint
	}
)

// Add will add an endpoint to the handler.
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

// ServeHTTP matches the request to the appropriate route and calls its handler.
func (h *EndPointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, endpoint := range h.endpoints {
		if endpoint.Regex.MatchString(r.URL.Path) && endpoint.Method == r.Method {
			endpoint.Handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

// Start loads the web server and begins listening.
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

	cleanupTicker := time.Tick(1 * time.Hour)
	go func() {
		for {
			select {
			case <-cleanupTicker:
				err := cleanupFiles()
				if err != nil {
					log.Warn(err.Error())
				}
			}
		}
	}()

	endpoints := EndPointHandler{}
	svr := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
	}

	endpoints.Add("GET", `/upgrade`, sendConduitBinary)
	endpoints.Add("POST", `/get`, getMessage)
	endpoints.Add("POST", "/put", putMessage)
	endpoints.Add("POST", "/stats/clients", clientStats)
	endpoints.Add("POST", "/stats", systemStats)
	endpoints.Add("POST", "/delete", deleteMessage)
	endpoints.Add("POST", "/deploy/list", deployInfo)
	endpoints.Add("POST", "/deploy/respond", deployRespond)
	endpoints.Add("POST", "/register", register)
	endpoints.Add("POST", "/deregister", deregister)
	endpoints.Add("POST", "/upload", acceptFile)
	endpoints.Add("POST", "/checkfile", checkfile)
	endpoints.Add("POST", "/asset", getAsset)
	http.Handle("/", &endpoints)
	serverRunning = true
	err := svr.ListenAndServe()
	return err
}

// writeResponse is a helper method that writes arbitrary structures in JSON.
// The interface passed to this mehtod should be a response structure form the
// postmaster/api package.
func writeResponse(w *http.ResponseWriter, response interface{}) error {
	bytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	io.WriteString(*w, string(bytes))
	return nil
}

// sendError is a helper method used to write out error responses. Any message
// passed to it will be wrapped in an error response and written out.
func sendError(w http.ResponseWriter, msg string) {
	e := &api.ApiError{
		Error: msg,
	}
	log.Error(msg)
	w.WriteHeader(http.StatusBadRequest)
	writeResponse(&w, e)
}

// readRequest is a helper method to read the request body and convert it into
// an arbitrary structure. The structure passed to it should be one of the
// requests from the postmaster/api package.
func readRequest(r *http.Request, req interface{}) error {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, req)
	return err
}

// sendConduitBinary will transmit the application's binary. It is used by
// clients to perform version syncs.
func sendConduitBinary(w http.ResponseWriter, r *http.Request) {
	log.Info("Upgrade requested")
	w.Header().Set("Content-Disposition", "attachment; filename=conduit")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	defer r.Body.Close()

	exePath, _ := osext.Executable()
	http.ServeFile(w, r, exePath)
}

func cleanupFiles() error {
	files, _ := ioutil.ReadDir(filesPath())
	for _, f := range files {
		pending, err := mailbox.AssetPending(f.Name())
		if err == nil && !pending {
			log.Infof("Cleaning up file %s", f.Name())
			err := os.Remove(filepath.Join(filesPath(), f.Name()))
			if err != nil {
				log.Warn("File clenaup " + err.Error())
			}
		} else if err != nil {
			return err
		} else {
			if time.Since(f.ModTime()) > 720*time.Hour {
				err := os.Remove(filepath.Join(filesPath(), f.Name()))
				if err != nil {
					log.Warn("File clenaup " + err.Error())
				}
			}
		}
	}
	return nil
}
