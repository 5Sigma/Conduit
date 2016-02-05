package engine

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"postmaster/client"
)

type ScriptEngine struct {
	VM        *otto.Otto
	Constants map[string]string
}

//read javascript file and execute
func (eng *ScriptEngine) ExecuteFile(filepath string) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("Error reading file: ", err.Error())
		os.Exit(-1)
	}
	err = eng.Execute(string(data))
	if err != nil {
		fmt.Println(err)
	}
}

//executes commands
func (eng *ScriptEngine) Execute(script string) error {
	_, err := eng.VM.Run(script)
	return err
}

func jsThrow(call otto.FunctionCall, err error) {
	value, _ := call.Otto.Call("new Error", nil, err.Error())
	panic(value)
}

func (eng *ScriptEngine) Constant(name, value string) {
	eng.VM.Set(name, value)
	eng.Constants[name] = value
}

func New() *ScriptEngine {
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
	fileObj.Set("eachFile", _file_eachFile)

	requestObj, _ := vm.Object(`$http = {}`)
	requestObj.Set("download", _http_download)
	requestObj.Set("get", _http_get)
	requestObj.Set("post", _http_post)

	systemObj, _ := vm.Object(`$system = {}`)
	systemObj.Set("executeAndRead", _system_executeAndRead)
	systemObj.Set("execute", _system_execute)
	systemObj.Set("detach", _system_detach)
	systemObj.Set("currentUser", _system_currentUser)
	systemObj.Set("kill", _system_kill)
	systemObj.Set("env", _system_env)
	systemObj.Set("expand", _system_expand)

	zipObj, _ := vm.Object(`$zip = {}`)
	zipObj.Set("compress", _zip_compress)
	zipObj.Set("decompress", _zip_decompress)

	vm.Set("$", _respond)
	eng := &ScriptEngine{VM: vm}
	eng.Constants = make(map[string]string)
	return eng
}

func getConstant(vm *otto.Otto, name string) string {
	val, err := vm.Get(name)
	if err != nil {
		return ""
	}
	str, err := val.ToString()
	if err != nil {
		return ""
	}
	return str
}

func _respond(call otto.FunctionCall) otto.Value {
	response, _ := call.Argument(0).ToString()
	client := client.Client{
		Host:  viper.GetString("queue.host"),
		Token: viper.GetString("access_key"),
	}
	messageId := getConstant(call.Otto, "SCRIPT_ID")
	err := client.Respond(messageId, response)
	if err != nil {
		jsThrow(call, err)
	}
	var v otto.Value
	v, _ = otto.ToValue(true)
	return v
}
