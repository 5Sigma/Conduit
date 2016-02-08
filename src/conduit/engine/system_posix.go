// +build !windows

package engine

import (
	"github.com/robertkrimen/otto"
	"os/exec"
	"syscall"
)

func _system_kill(call otto.FunctionCall) otto.Value {
	processName, _ := call.Argument(0).ToString()
	commandStrings := []string{processName}
	if err := exec.Command("pkill", commandStrings...).Run(); err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}

//executes a program outside of the conduit process group
func _system_detach(call otto.FunctionCall) otto.Value {
	cmdName, _ := call.Argument(0).ToString()
	cmdArgs, _ := call.Argument(1).Export()
	commandStrings := []string{}
	if strings, ok := cmdArgs.([]string); ok {
		commandStrings = strings
	}
	cmd := exec.Command(cmdName, commandStrings...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}

func _system_shell(call otto.FunctionCall) otto.Value {
	var (
		cmdOut []byte
		err    error
	)
	cmd, _ := call.Argument(0).ToString()
	if cmdOut, err = exec.Command("/bin/sh", "-c", cmd).Output(); err != nil {
		jsThrow(call, err)
	}
	v, _ := otto.ToValue(string(cmdOut))
	return v
}
