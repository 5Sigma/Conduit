package engine

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"os"
)

func ExecuteFile(filepath string) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("Error reading file: ", err.Error())
		os.Exit(-1)
	}
	Execute(string(data))
}

func Execute(script string) error {
	vm := createVM()
	_, err := vm.Run(script)
	return err
}

func createVM() *otto.Otto {
	vm := otto.New()

	logObj, _ := vm.Object(`$log = {}`)
	logObj.Set("info", _log_info)

	fileObj, _ := vm.Object(`$file = {}`)
	fileObj.Set("exists", _file_exists)

	return vm
}

func _log_info(call otto.FunctionCall) otto.Value {
	msg, _ := call.Argument(0).ToString()
	fmt.Println(msg)
	return otto.Value{}
}

func _file_exists(call otto.FunctionCall) otto.Value {
	filepath, _ := call.Argument(0).ToString()
	var val otto.Value
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		val, _ = otto.ToValue(false)
	} else {
		val, _ = otto.ToValue(true)
	}
	return val
}
