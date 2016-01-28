package engine

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"os/exec"
	"os/user"
	"syscall"
)

//executes a program
func _system_execute(call otto.FunctionCall) otto.Value {
	cmdName, _ := call.Argument(0).ToString()
	cmdArgs, _ := call.Argument(1).Export()
	throw, _ := call.Argument(2).ToString()
	commandStrings := []string{}
	if strings, ok := cmdArgs.([]string); ok {
		commandStrings = strings
	}
	if err := exec.Command(cmdName, commandStrings...).Run(); err != nil {
		if throw != "false" {
			jsThrow(call, err)
		}
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

//executes and returns value
func _system_executeAndRead(call otto.FunctionCall) otto.Value {
	var (
		cmdOut []byte
		err    error
	)
	cmdName, _ := call.Argument(0).ToString()
	cmdArgs, _ := call.Argument(1).Export()
	commandStrings := []string{}
	if strings, ok := cmdArgs.([]string); ok {
		commandStrings = strings
	}
	if cmdOut, err = exec.Command(cmdName, commandStrings...).Output(); err != nil {
		jsThrow(call, err)
	}
	v, _ := otto.ToValue(string(cmdOut))
	return v
}

//returns information about the system
func _system_currentUser(call otto.FunctionCall) otto.Value {
	currentUser, err := user.Current()
	if err != nil {
		jsThrow(call, err)
	}
	objString := fmt.Sprintf(`
		({
			name: '%s',
			homeDir: '%s',
			uid: '%s',
			gid: '%s',
			username: '%s'
		})
	`, currentUser.Name, currentUser.HomeDir, currentUser.Uid, currentUser.Gid,
		currentUser.Username)
	obj, err := call.Otto.Object(objString)
	if err != nil {
		jsThrow(call, err)
	}
	v, err := otto.ToValue(obj)
	if err != nil {
		jsThrow(call, err)
	}
	return v
}
