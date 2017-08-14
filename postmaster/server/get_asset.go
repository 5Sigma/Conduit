package server

import (
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/api"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"net/http"
	"os"
	"path/filepath"
)

func getAsset(w http.ResponseWriter, r *http.Request) {
	var (
		request = api.GetAssetRequest{}
	)

	err := readRequest(r, &request)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if accessKey == nil {
		sendError(w, "Access key is invalid")
		return
	}

	if !request.Validate(accessKey.Secret) {
		sendError(w, "Could not validate signature")
		return
	}

	fp := filepath.Join(filesPath(), request.MD5)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		sendError(w, "Asset not found on server")
		return
	}

	log.Infof("Serving asset to %s", accessKey.Name)
	http.ServeFile(w, r, fp)

}
