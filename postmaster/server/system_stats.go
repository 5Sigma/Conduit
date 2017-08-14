package server

import (
	"github.com/5sigma/conduit/postmaster/api"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"io/ioutil"
	"net/http"
	"runtime"
)

// systemStats is json endpoint used to retrieve overall statistics. The token
// used to request the endpoint must have write priviledges.
func systemStats(w http.ResponseWriter, r *http.Request) {
	var (
		request  api.SimpleRequest
		memStats runtime.MemStats
	)
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not get stats")
		return
	}

	accessKey, _ := mailbox.FindKeyByName(request.AccessKeyName)

	if accessKey == nil {
		sendError(w, "Access key is invalid")
		return
	}

	if !accessKey.CanAdmin() {
		sendError(w, "Not allowed to get statistics")
		return
	}

	if !request.Validate(accessKey.Secret) {
		sendError(w, "Could not validate signature")
		return
	}

	stats, err := mailbox.Stats()
	if err != nil {
		sendError(w, err.Error())
		return
	}

	dbVersion, err := mailbox.GetDBVersion()
	if err != nil {
		sendError(w, err.Error())
		return
	}

	var fileCount int64 = 0
	var filesSize int64 = 0
	files, _ := ioutil.ReadDir(filesPath())
	for _, f := range files {
		fileCount++
		filesSize += f.Size()
	}

	runtime.ReadMemStats(&memStats)

	response := api.SystemStatsResponse{
		TotalMailboxes:   stats.MailboxCount,
		PendingMessages:  stats.PendingMessages,
		MessageCount:     stats.MessageCount,
		ConnectedClients: int64(len(pollingChannels)),
		DBVersion:        dbVersion,
		CPUCount:         int64(runtime.NumCPU()),
		Threads:          int64(runtime.NumGoroutine()),
		MemoryAllocated:  memStats.Alloc,
		Lookups:          memStats.Lookups,
		NextGC:           memStats.NextGC,
		FileStoreSize:    filesSize,
		FileStoreCount:   fileCount,
	}
	response.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, response)
}
