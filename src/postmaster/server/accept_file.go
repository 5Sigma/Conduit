package server

import (
	"conduit/log"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/kardianos/osext"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"postmaster/api"
	"postmaster/mailbox"
)

func acceptFile(w http.ResponseWriter, r *http.Request) {
	var (
		request = api.UploadFileRequest{}
	)
	defer r.Body.Close()
	// parse form post data
	err := r.ParseMultipartForm(1000000)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	// Get reuqest data and unmarshal
	reqData := r.FormValue("data")
	err = json.Unmarshal([]byte(reqData), &request)
	if err != nil {
		sendError(w, "Could not parse request!")
		return
	}
	accessKey, _ := mailbox.FindKeyByName(request.AccessKeyName)
	if accessKey == nil {
		sendError(w, "Invalid access key")
		return
	}
	if !accessKey.CanAdmin() {
		sendError(w, "Not allowed to upload files")
		return
	}
	if !request.Validate(accessKey.Secret) {
		sendError(w, "Invalid signature")
		return
	}

	// Save the posted file
	file, _, err := r.FormFile("file")
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		sendError(w, err.Error())
		return
	}

	path := filepath.Join(filesPath(), request.MD5)
	out, err := os.Create(path)
	if out != nil {
		defer out.Close()
	}
	if err != nil {
		sendError(w, err.Error())
		return
	}
	_, err = io.Copy(out, file)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	fileHash, err := hashFile(path)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	if request.MD5 != fileHash {
		defer os.Remove(filepath.Join(filesPath(), request.MD5))
		sendError(w, fmt.Sprintf("MD5 missmatch %s != %s", request.MD5, fileHash))
		return
	}

	log.Infof("File uploaded %s", request.MD5)
	response := api.SimpleResponse{Success: true}
	response.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, response)
}

func filesPath() string {
	binPath, _ := osext.ExecutableFolder()
	path := filepath.Join(binPath, "files")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
	return path
}

func hashFile(fp string) (string, error) {
	file, err := os.Open(fp)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		return "", err
	}
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashStr := fmt.Sprintf("%x", hash.Sum(nil))
	return hashStr, nil
}

func checkfile(w http.ResponseWriter, r *http.Request) {
	req := api.CheckFileRequest{}
	err := readRequest(r, &req)
	if err != nil {
		sendError(w, "Could not parse request")
		return
	}

	accessKey, _ := mailbox.FindKeyByName(req.AccessKeyName)
	if accessKey == nil {
		sendError(w, "Access key invalid")
		return
	}

	path := filepath.Join(filesPath(), req.MD5)

	resp := api.SimpleResponse{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		resp.Success = false
	} else {
		resp.Success = true
	}
	resp.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, resp)
}
