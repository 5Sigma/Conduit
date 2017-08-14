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

func _system_shell(call otto.FunctionCall) otto.Value {
	var (
		cmdOut []byte
		err    error
	)
	cmd, _ := call.Argument(0).ToString()
	if cmdOut, err = exec.Command("cmd.exe", "/c", cmd).Output(); err != nil {
		jsThrow(call, err)
	}
	v, _ := otto.ToValue(string(cmdOut))
	return v
}
