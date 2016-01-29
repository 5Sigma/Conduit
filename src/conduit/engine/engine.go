package engine

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"os"
)

//read javascript file and execute
func ExecuteFile(filepath string) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("Error reading file: ", err.Error())
		os.Exit(-1)
	}
	err = Execute(string(data))
	if err != nil {
		fmt.Println(err)
	}
}

//executes commands
func Execute(script string) error {
	vm := createVM()
	_, err := vm.Run(script)
	return err
}

func jsThrow(call otto.FunctionCall, err error) {
	value, _ := call.Otto.Call("new Error", nil, err.Error())
	panic(value)
}

func createVM() *otto.Otto {
	vm := otto.New()

	logObj, _ := vm.Object(`$log = {}`)
	logObj.Set("info", _log_info)

	fileObj, _ := vm.Object(`$file = {}`)
	fileObj.Set("exists", _file_exists)
	fileObj.Set("write", _file_write)
	fileObj.Set("copy", _file_copy)
	fileObj.Set("size", _file_size)
	fileObj.Set("move", _file_move)
	fileObj.Set("mkdir", _file_mkdir)
	fileObj.Set("delete", _file_delete)
	fileObj.Set("readString", _file_readString)

	requestObj, _ := vm.Object(`$request = {}`)
	requestObj.Set("download", _request_download)
	requestObj.Set("get", _request_get)

	systemObj, _ := vm.Object(`$system = {}`)
	systemObj.Set("executeAndRead", _system_executeAndRead)
	systemObj.Set("execute", _system_execute)
	systemObj.Set("detach", _system_detach)
	systemObj.Set("currentUser", _system_currentUser)
	systemObj.Set("kill", _system_kill)

	return vm
}
