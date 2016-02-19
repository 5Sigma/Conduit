package engine

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	defer file.Close()
	if err != nil {
		jsThrow(call, err)
	}

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

	if !fileExists(sourcePath) {
		jsThrow(call, errors.New("Source file does not exist"))
	}

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

	if !fileExists(sourcePath) {
		jsThrow(call, errors.New("Source file not found "+sourcePath))
	}

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

//_file_info gathers path information.
func _file_info(call otto.FunctionCall) otto.Value {
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

	obj, err := call.Otto.Object(`new Object()`)

	if err != nil {
		jsThrow(call, err)
	}

	size, _ := otto.ToValue(info.Size())

	obj.Set("size", size)

	modTime, err := otto.ToValue(info.ModTime().Unix())
	if err != nil {
		jsThrow(call, err)
	}
	obj.Set("lastModified", modTime)

	name, _ := otto.ToValue(info.Name())
	obj.Set("name", name)

	isDir, _ := otto.ToValue(info.IsDir())
	obj.Set("isDir", isDir)

	objV, _ := otto.ToValue(obj)
	return objV
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

func _file_tempFile(call otto.FunctionCall) otto.Value {
	f, err := ioutil.TempFile("", "conduit")
	if err != nil {
		jsThrow(call, err)
	}
	defer f.Close()
	v, _ := otto.ToValue(f.Name())
	return v
}

func _file_tempFolder(call otto.FunctionCall) otto.Value {
	d, err := ioutil.TempDir("", "conduit")
	if err != nil {
		jsThrow(call, err)
	}
	v, _ := otto.ToValue(d)
	return v
}

func _file_join(call otto.FunctionCall) otto.Value {
	paths := []string{}
	for idx, _ := range call.ArgumentList {
		str, _ := call.Argument(idx).ToString()
		paths = append(paths, str)
	}
	pathStr := filepath.Join(paths...)
	v, _ := otto.ToValue(pathStr)
	return v
}

func _file_base(call otto.FunctionCall) otto.Value {
	fullPath, _ := call.Argument(0).ToString()
	v, _ := otto.ToValue(path.Base(fullPath))
	return v
}

func _file_cleanPath(call otto.FunctionCall) otto.Value {
	fullPath, _ := call.Argument(0).ToString()
	v, _ := otto.ToValue(path.Clean(fullPath))
	return v
}

func _file_dir(call otto.FunctionCall) otto.Value {
	fullPath, _ := call.Argument(0).ToString()
	v, _ := otto.ToValue(path.Dir(fullPath))
	return v
}

func _file_ext(call otto.FunctionCall) otto.Value {
	fullPath, _ := call.Argument(0).ToString()
	v, _ := otto.ToValue(path.Ext(fullPath))
	return v
}

func _file_eachLine(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 2 {
		jsThrow(call, errors.New("Wrong number of arguments."))
	}

	sourcePath, _ := call.Argument(0).ToString()
	fn := call.Argument(1)
	if !fileExists(sourcePath) {
		jsThrow(call, errors.New("Source file doesn't exist"))
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		jsThrow(call, err)
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			v, _ := otto.ToValue(line)
			fn.Call(v, v)
		}
	}

	if err := scanner.Err(); err != nil {
		jsThrow(call, err)
	}

	return otto.Value{}
}

func _file_md5(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) == 0 {
		jsThrow(call, errors.New("Invalid arguments"))
	}
	fp, _ := call.Argument(0).ToString()
	file, err := os.Open(fp)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		jsThrow(call, err)
	}
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		jsThrow(call, err)
	}
	hashStr := fmt.Sprintf("%x", hash.Sum(nil))
	v, _ := otto.ToValue(hashStr)
	return v
}
