package engine

import (
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// fileExists is a helper function used by other file functions in the package.
func fileExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

//EXISTS
//checks if file exists returning true or false
func _file_exists(call otto.FunctionCall) otto.Value {
	filepath, _ := call.Argument(0).ToString()
	v, _ := otto.ToValue(fileExists(filepath))
	return v
}

//WRITE
//writes string to a file, creating the file if it does not exist
func _file_write(call otto.FunctionCall) otto.Value {
	filepath, _ := call.Argument(0).ToString()
	data, _ := call.Argument(1).ToString()
	if !fileExists(filepath) {
		os.Create(filepath)
	}

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		jsThrow(call, err)
	}

	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}

//COPY
//copies a file, overwriting the destination if exists.
func _file_copy(call otto.FunctionCall) otto.Value {
	sourcePath, _ := call.Argument(0).ToString()
	destinationPath, _ := call.Argument(1).ToString()

	//check if destination exists and delete if so
	if fileExists(destinationPath) {
		if err := os.Remove(destinationPath); err != nil {
			jsThrow(call, err)
		}
	}

	//read source file
	in, err := os.Open(sourcePath)
	if err != nil {
		jsThrow(call, err)
	}
	defer in.Close()

	//create destination file
	out, err := os.Create(destinationPath)
	if err != nil {
		jsThrow(call, err)
	}
	defer out.Close()

	//copy contents of source to destination
	_, err = io.Copy(out, in)
	_ = out.Close()
	if err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}

//MOVE
//moves a file, overwriting the destination if exists
func _file_move(call otto.FunctionCall) otto.Value {
	sourcePath, _ := call.Argument(0).ToString()
	destinationPath, _ := call.Argument(1).ToString()

	//check if destination exists and delete if so
	if fileExists(destinationPath) {
		if err := os.Remove(destinationPath); err != nil {
			jsThrow(call, err)
		}
	}

	err := os.Rename(sourcePath, destinationPath)
	if err != nil {
		jsThrow(call, err)
	}

	return otto.Value{}
}

//SIZE
//returns size for give file
func _file_size(call otto.FunctionCall) otto.Value {
	filepath, _ := call.Argument(0).ToString()

	file, err := os.Open(filepath)
	if err != nil {
		jsThrow(call, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		jsThrow(call, err)
	}
	v, _ := otto.ToValue(info.Size())
	return v
}

//MKDIR
//creates a directory, along with any necessary parents
func _file_mkdir(call otto.FunctionCall) otto.Value {
	path, _ := call.Argument(0).ToString()

	err := os.MkdirAll(path, 0755)
	if err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}

//DELETE
//deletes path/file and any children it contains
func _file_delete(call otto.FunctionCall) otto.Value {
	path, _ := call.Argument(0).ToString()

	err := os.RemoveAll(path)
	if err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}

//READSTRING
//returns the data within the specified file
func _file_readString(call otto.FunctionCall) otto.Value {
	filepath, _ := call.Argument(0).ToString()

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		jsThrow(call, err)
	}
	v, _ := otto.ToValue(string(data))
	return v
}

// _file_eachFile finds every file that matches a pattern and calls a passed
// function for each file path.
func _file_eachFile(call otto.FunctionCall) otto.Value {
	pattern, _ := call.Argument(0).ToString()
	f := call.Argument(1)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		jsThrow(call, err)
	}
	for _, m := range matches {
		om, _ := otto.ToValue(m)
		f.Call(om, om)
	}
	return otto.Value{}
}
