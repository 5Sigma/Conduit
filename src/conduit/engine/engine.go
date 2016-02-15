package engine

import (
	"conduit/info"
	"fmt"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"postmaster/client"
	"runtime"
)

type (
	ScriptEngine struct {
		VM        *otto.Otto
		Constants map[string]string
		AssetPath string
	}
)

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

func (eng *ScriptEngine) Validate(script string) error {
	_, err := parser.ParseFile(nil, "", script, 0)
	return err
}

func (eng *ScriptEngine) GetVar(name, script string) (interface{}, error) {
	program, err := parser.ParseFile(nil, "", script, 0)
	if err != nil {
		return nil, err
	}
	for _, node := range program.Body {
		if stmt, ok := node.(*ast.VariableStatement); ok {
			if len(stmt.List) == 0 {
				continue
			}
			if exp, ok := stmt.List[0].(*ast.VariableExpression); ok {
				if literal, ok := exp.Initializer.(*ast.NumberLiteral); ok {
					return literal.Value, nil
				}
				if literal, ok := exp.Initializer.(*ast.StringLiteral); ok {
					return literal.Value, nil
				}
			}
		}
	}
	return nil, nil
}
func getFunctionLiteral(name string, script string) (string, error) {
	program, err := parser.ParseFile(nil, "", script, 0)
	if err != nil {
		return "", err
	}
	for _, node := range program.Body {
		if stmt, ok := node.(*ast.ExpressionStatement); ok {
			if exp, ok := stmt.Expression.(*ast.AssignExpression); ok {
				if left, ok := exp.Left.(*ast.Identifier); ok {
					if _, ok := exp.Right.(*ast.FunctionLiteral); ok {
						if left.Name == name {
							return script[stmt.Idx0()-1 : stmt.Idx1()], nil
						}
					}
				}
			}
		}
	}
	return fmt.Sprintf("%s = function() {};", name), nil
}

func (eng *ScriptEngine) ExecuteFunction(name, script string) (string, error) {
	fStr, err := getFunctionLiteral(name, script)
	fScript := fmt.Sprintf("%s; var result = %s();", fStr, name)
	_, err = eng.VM.Run(fScript)
	if err != nil {
		return "", err
	}
	res, err := eng.VM.Get("result")
	if err != nil {
		return "", err
	}
	resStr, err := res.ToString()
	if err != nil {
		return "", err
	}
	return resStr, nil
}

func New() *ScriptEngine {
	vm := otto.New()

	logObj, _ := vm.Object(`$log = {}`)
	logObj.Set("info", _log_info)

	fileObj, _ := vm.Object(`$file = {}`)
	fileObj.Set("exists", _file_exists)
	fileObj.Set("write", _file_write)
	fileObj.Set("info", _file_info)
	fileObj.Set("copy", _file_copy)
	fileObj.Set("size", _file_size)
	fileObj.Set("move", _file_move)
	fileObj.Set("mkdir", _file_mkdir)
	fileObj.Set("delete", _file_delete)
	fileObj.Set("readString", _file_readString)
	fileObj.Set("eachFile", _file_eachFile)
	fileObj.Set("tempFile", _file_tempFile)
	fileObj.Set("tempFolder", _file_tempFolder)
	fileObj.Set("join", _file_join)

	requestObj, _ := vm.Object(`$http = {}`)
	requestObj.Set("download", _http_download)
	requestObj.Set("get", _http_get)
	requestObj.Set("post", _http_post)

	systemObj, _ := vm.Object(`$system = {}`)
	systemObj.Set("shell", _system_shell)
	systemObj.Set("executeAndRead", _system_executeAndRead)
	systemObj.Set("execute", _system_execute)
	systemObj.Set("detach", _system_detach)
	systemObj.Set("currentUser", _system_currentUser)
	systemObj.Set("kill", _system_kill)
	systemObj.Set("env", _system_env)
	systemObj.Set("expand", _system_expand)
	systemObj.Set("PLATFORM", runtime.GOOS)
	systemObj.Set("ARCH", runtime.GOARCH)
	systemObj.Set("VERSION", info.ConduitVersion)

	zipObj, _ := vm.Object(`$zip = {}`)
	zipObj.Set("compress", _zip_compress)
	zipObj.Set("decompress", _zip_decompress)

	tarObj, _ := vm.Object(`$tar = {}`)
	tarObj.Set("compress", _tar_compress)
	tarObj.Set("decompress", _tar_decompress)

	gzipObj, _ := vm.Object(`$gzip = {}`)
	gzipObj.Set("compress", _gzip_compress)
	gzipObj.Set("decompress", _gzip_decompress)

	assetObj, _ := vm.Object(`$asset = {}`)
	assetObj.Set("exists", _asset_exists)

	vm.Set("$", _respond)
	vm.Set("$agent", _agent)

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
		Host:          viper.GetString("host"),
		AccessKey:     viper.GetString("access_key"),
		AccessKeyName: viper.GetString("access_key_name"),
		ShowRequests:  viper.GetBool("show_requests"),
	}
	if client.AccessKeyName == "" {
		client.AccessKeyName = viper.GetString("mailbox")
	}
	messageId := getConstant(call.Otto, "SCRIPT_ID")
	err := client.Respond(messageId, response, false)
	if err != nil {
		jsThrow(call, err)
	}
	var v otto.Value
	v, _ = otto.ToValue(true)
	return v
}
