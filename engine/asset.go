package engine

import (
	"github.com/robertkrimen/otto"
)

func _asset_exists(call otto.FunctionCall) otto.Value {
	assetObj, _ := call.Otto.Object("$asset")
	filepathV, _ := assetObj.Get("filepath")
	if filepathV.IsUndefined() {
		v, _ := call.Otto.ToValue(false)
		return v
	}
	filepath := filepathV.String()
	if filepath == "" {
		v, _ := call.Otto.ToValue(false)
		return v
	}
	exists := fileExists(filepath)
	v, _ := call.Otto.ToValue(exists)
	return v

}

func (eng *ScriptEngine) SetAsset(path string) {
	assetObj, _ := eng.VM.Object("$asset")
	pathV, _ := eng.VM.ToValue(path)
	assetObj.Set("filepath", pathV)
}
