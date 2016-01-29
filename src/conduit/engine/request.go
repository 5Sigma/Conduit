package engine

import (
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func _request_download(call otto.FunctionCall) otto.Value {
	url, _ := call.Argument(0).ToString()
	filepath, _ := call.Argument(1).ToString()
	out, err := os.Create(filepath)
	if err != nil {
		jsThrow(call, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		jsThrow(call, err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		jsThrow(call, err)
	}

	return otto.Value{}
}

func _request_get(call otto.FunctionCall) otto.Value {
	url, _ := call.Argument(0).ToString()

	response, err := http.Get(url)
	if err != nil {
		jsThrow(call, err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		jsThrow(call, err)
	}
	v, err := otto.ToValue(string(content))
	if err != nil {
		jsThrow(call, err)
	}
	return v
}
