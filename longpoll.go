// handlers.go
package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// handle long poll directory request. ONLY FOR LISTED NODES!
func LPDirectoryHandler(w http.ResponseWriter, req *http.Request) {
	DLock.RLock()
	w.Header().Add("Content-Type", "text/plain")
	// Assure that the node is listed
	listed := false
	for i := 0; i < len(KDirectory); i++ {
		if strings.Split(KDirectory[i].Address, ":")[0] ==
			strings.Split(req.RemoteAddr, ":")[0] {
			listed = true
		}
	}
	DLock.RUnlock()
	if !listed {
		fmt.Fprintf(w, "Error: you are not a listed node (yet)\n")
		return
	}
	// Wait until directory change happens
	crn := revnum
	for revnum == crn {
		time.Sleep(time.Second / 10)
	}
	ReadDirectoryHandler(w, req)
}
