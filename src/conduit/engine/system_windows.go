// +build windows

package engine

import (
	"github.com/robertkrimen/otto"
	"os/exec"
)

func _system_kill(call otto.FunctionCall) otto.Value {
	processName, _ := call.Argument(0).ToString()
	commandStrings := []string{"/F", "/IM", processName}
	if err := exec.Command("taskkill", commandStrings...).Run(); err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}

func _system_detach(call otto.FunctionCall) otto.Value {
	return _system_execute(call)
}
