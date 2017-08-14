package agent

import (
	"encoding/json"
	"github.com/5sigma/conduit/engine"
	"github.com/5sigma/conduit/postmaster/api"
	"io"
	"io/ioutil"
	"net/http"
)

type (
	Agent struct {
		Address   string
		AccessKey string
	}
)

var (
	Address   string = "127.0.0.1:4112"
	AccessKey string
)

// Start will start the agent at the address agent.Address.
func Start() error {
	http.HandleFunc("/", command)
	return http.ListenAndServe(Address, nil)
}

// getRequest is a helper function to read the JSON data into the AgentRequest
// structure.
func getRequest(r *http.Request) (*api.AgentRequest, error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	req := &api.AgentRequest{}
	err = json.Unmarshal(data, req)
	return req, err
}

// writeResponse is a helper function which will write out an AgentReponse
// structure.
func writeResponse(w http.ResponseWriter, response *api.AgentResponse) error {
	bytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	if response.Success == false {
		w.WriteHeader(http.StatusBadRequest)
	}
	io.WriteString(w, string(bytes))
	return nil
}

// command will process any request to the server. It will validate the request
// and then process any script given.
func command(w http.ResponseWriter, r *http.Request) {
	request, err := getRequest(r)
	response := &api.AgentResponse{}

	if !request.Validate(AccessKey) {
		response.Success = false
		response.Error = "Could not validate signature"
		response.Sign("", AccessKey)
		writeResponse(w, response)
		return
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
		response.Sign("", AccessKey)
		writeResponse(w, response)
		return
	}

	if !request.Validate(AccessKey) {
		response.Success = false
		response.Error = "Invalid signature"
		response.Sign("", AccessKey)
		writeResponse(w, response)
		return
	}

	eng := engine.New()
	err = eng.Execute(request.Function)
	if err != nil {
		response.Success = false
		response.Error = err.Error()
		response.Sign("", AccessKey)
		writeResponse(w, response)
		return
	}

	response.Success = true
	response.Sign("", AccessKey)
	writeResponse(w, response)
}
