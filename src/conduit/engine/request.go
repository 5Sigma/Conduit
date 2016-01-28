package engine

import (
	"github.com/robertkrimen/otto"
	"io"
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
