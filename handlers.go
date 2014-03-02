// handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"code.google.com/p/log4go"

	"code.google.com/p/go.crypto/openpgp/clearsign"
)

// handle read directory request
func ReadDirectoryHandler(w http.ResponseWriter, req *http.Request) {
	rhost := req.Header.Get("CF-Connecting-IP")
	log4go.Info("Read request from %s", rhost)
	DLock.RLock()
	defer DLock.RUnlock()
	w.Header().Add("Content-Type", "text/plain")
	// make our own kdirectory
	oAdj := GetAdjacentNodes(rhost)
	for i := 0; i < len(KDirectory); i++ {
		if strings.Split(KDirectory[i].Address, ":")[0] == strings.Split(rhost, ":")[0] {
			oAdj = KDirectory[i].Adjacents
		}
	}
	fDir := make([]KNode, len(KDirectory))
	for i := 0; i < len(fDir); i++ {
		fDir[i] = KDirectory[i]
		contains := false
		for j := 0; j < len(oAdj); j++ {
			contains = contains || i == oAdj[j]
		}
		if !contains {
			fDir[i].Address = "(hidden)"
		}
	}
	nW, _ := clearsign.Encode(w, OurRSAKey.PrivateKey, nil)
	defer nW.Close()
	b, _ := json.MarshalIndent(fDir, "", "  ")
	nW.Write(b)
}

func RFormatDirectoryHandler(w http.ResponseWriter, req *http.Request) {
	DLock.RLock()
	defer DLock.RUnlock()
	fmt.Fprintf(w, "library(igraph)\n")
	fmt.Fprintf(w, "adjlist <- list()\n")
	for i := 0; i < len(KDirectory); i++ {
		fmt.Fprintf(w, "adjlist <- append(adjlist, list(c(")
		for j := 0; j < len(KDirectory[i].Adjacents)-1; j++ {
			fmt.Fprintf(w, "%d,", KDirectory[i].Adjacents[j]+1)
		}
		j := len(KDirectory[i].Adjacents) - 1
		fmt.Fprintf(w, "%d", KDirectory[i].Adjacents[j]+1)
		fmt.Fprintf(w, ")))")
		if i < len(KDirectory)-1 {
			fmt.Fprintf(w, "\n")
		}
	}
	fmt.Fprintf(w, "\ntkplot(graph.adjlist(adjlist, mode='all'))\n")
}

// handle upload info request
func UploadInfoHandler(w http.ResponseWriter, req *http.Request) {
	rhost := req.Header.Get("CF-Connecting-IP")
	req.ParseForm()
	w.Header().Add("Content-Type", "text/plain")
	theirport := req.Form.Get("port")
	theirprotocol := req.Form.Get("protocol")
	theirpkey := req.Form.Get("keyhash")
	theirhost := strings.Join([]string{rhost, theirport}, "")
	realprotoc, err := strconv.Atoi(theirprotocol)
	if err != nil {
		fmt.Fprintf(w, "Error encountered while uploading info:\n%s\n", err.Error())
		return
	}
	//TODO: VERIFY THE DATA!!!
	AddNode(theirhost, theirpkey, realprotoc)
}
