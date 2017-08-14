package engine

import (
	"fmt"
	"github.com/robertkrimen/otto"
)

func _log_info(call otto.FunctionCall) otto.Value {
	msg, _ := call.Argument(0).ToString()
	fmt.Println(msg)
	return otto.Value{}
}
