package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"net/http"
	"postmaster/api"
)

var (
	Agents         = make(map[string]string)
	AgentAccessKey string
)

func _agent(call otto.FunctionCall) otto.Value {
	agentName, _ := call.Argument(0).ToString()
	fn, _ := call.Argument(1).ToString()

	var address string
	if a, ok := Agents[agentName]; ok {
		address = a
	} else {
		jsThrow(call, errors.New("agent not found"))
	}

	req := &api.AgentRequest{
		Function: fmt.Sprintf("(%s)();", fn),
	}
	req.Sign("", AgentAccessKey)
	http.DefaultClient.Timeout = 0
	requestBytes, err := json.Marshal(req)
	url := fmt.Sprintf("http://%s", address)
	reader := bytes.NewReader(requestBytes)
	resp, err := http.Post(url, "application/json", reader)
	if err != nil {
		jsThrow(call, err)
	}

	responseData, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		jsThrow(call, errors.New("Agent not found"))
	}

	if resp.StatusCode != 200 {
		var errorResponse api.ApiError
		json.Unmarshal(responseData, &errorResponse)
		jsThrow(call, errors.New(errorResponse.Error))
	}

	var response = api.AgentResponse{}
	json.Unmarshal(responseData, &response)

	return otto.Value{}
}
