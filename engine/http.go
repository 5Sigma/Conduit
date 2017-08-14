package engine

import (
	"bytes"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func _http_download(call otto.FunctionCall) otto.Value {
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

func _http_get(call otto.FunctionCall) otto.Value {
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

func _http_post(call otto.FunctionCall) otto.Value {
	url, _ := call.Argument(0).ToString()
	data, _ := call.Argument(1).ToString()
	contentType, _ := call.Argument(2).ToString()
	response, err := http.Post(url, contentType, bytes.NewBufferString(data))
	if err != nil {
		jsThrow(call, err)
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		jsThrow(call, err)
	}
	v, err := otto.ToValue(string(content))
	return v
}
