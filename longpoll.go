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
	rhost := req.Header.Get("CF-Connecting-IP")
	DLock.RLock()
	w.Header().Add("Content-Type", "text/plain")
	// Assure that the node is listed
	listed := false
	for i := 0; i < len(KDirectory); i++ {
		if strings.Split(KDirectory[i].Address, ":")[0] ==
			strings.Split(rhost, ":")[0] {
			listed = true
		}
	}
	DLock.RUnlock()
	if !listed {
		fmt.Fprintf(w, "Error: you are not a listed node (yet)\n")
		return
	}
	// Wait until directory change happens
	count := 100
	crn := revnum
	for revnum == crn {
		time.Sleep(time.Second / 10)
		count--
		if count == 0 {
			break
		}
	}
	ReadDirectoryHandler(w, req)
}
