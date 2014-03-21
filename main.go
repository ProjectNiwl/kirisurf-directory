// testdirectory project main.go
package main

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go.crypto/openpgp"
	"code.google.com/p/go.crypto/openpgp/packet"
	"code.google.com/p/log4go"
)

var revnum = 0

// Maintains the big database
type KNode struct {
	PublicKey       string
	Address         string
	ProtocolVersion int
	Adjacents       []int
	ExitNode        bool
}

// Mah kee
var OurRSAKey *openpgp.Entity

// The database itself
var KDirectory = make([]KNode, 0)

// Global lock for the database
var DLock sync.RWMutex

// Get the adjacent nodes of an address
func GetAdjacentNodes(addr string) []int {
	if len(KDirectory) == 0 {
		return []int{}
	}
	addr = strings.Split(addr, ":")[0]
	index := crc32.Checksum([]byte(addr), crc32.IEEETable)
	toret := make([]int, 0)
	for i := 0; i < 3; i++ {
		ptr := KDirectory[int((index+uint32(i))%uint32(len(KDirectory)))]
		if strings.Split(ptr.Address, ":")[0] != addr {
			toret = append(toret, int((index+uint32(i))%uint32(len(KDirectory))))
		}
	}
	return toret
}

// Add a node to the database
func AddNode(addr string, pkey string, pvers int, isexit bool) {
	log4go.Info("Adding %s with pkey %s and pvers %d", addr, pkey, pvers)
	DLock.Lock()
	defer func() {
		DLock.Unlock()
		revnum++
	}()
	adj := GetAdjacentNodes(addr)
	toadd := KNode{pkey, addr, pvers, adj, isexit}
	KDirectory = append(KDirectory, toadd)
	/*for i := 0; i < len(KDirectory); i++ {
		nd := &KDirectory[i]
		nd.Adjacents = GetAdjacentNodes(nd.Address)
	}*/
	// Now add the other direction
	for i := 0; i < len(KDirectory); i++ {
		lst := KDirectory[i].Adjacents
		for j := 0; j < len(lst); j++ {
			nd := &KDirectory[lst[j]]
			nd.Adjacents = append(nd.Adjacents, i)
		}
	}
	for i := 0; i < len(KDirectory); i++ {
		nd := &KDirectory[i]
		nd.Adjacents = FixDuplicates(nd.Adjacents)
	}
}

// Deletes a node from the database
func DeleteNode(idx int) {
	DLock.Lock()
	defer func() {
		DLock.Unlock()
	}()
	// new directory altogether
	olddir := KDirectory
	KDirectory = make([]KNode, len(olddir)-1)
	for i := 0; i < idx; i++ {
		KDirectory[i] = olddir[i]
	}
	for i := idx; i < len(KDirectory); i++ {
		KDirectory[i] = olddir[i+1]
	}
	// remove all references to the old index
	for i := 0; i < len(KDirectory); i++ {
		newadj := make([]int, 0)
		for j := 0; j < len(KDirectory[i].Adjacents); j++ {
			if KDirectory[i].Adjacents[j] != idx {
				newadj = append(newadj, KDirectory[i].Adjacents[j])
			}
		}
		KDirectory[i].Adjacents = newadj
	}
	// update all references
	for i := 0; i < len(KDirectory); i++ {
		for j := 0; j < len(KDirectory[i].Adjacents); j++ {
			if KDirectory[i].Adjacents[j] >= idx {
				KDirectory[i].Adjacents[j]--
			}
		}
	}
}

// Fixes duplicates
func FixDuplicates(thing []int) []int {
	log4go.Debug(thing)
	toret := make([]int, 0)
	for i := 0; i < len(thing); i++ {
		blah := false
		for j := 0; j < len(toret); j++ {
			if toret[j] == thing[i] {
				blah = true
			}
		}
		if !blah {
			toret = append(toret, thing[i])
		}
	}
	log4go.Debug(toret)
	return toret
}

// Prints directory in R form
func PrintDirectoryR() {
	DLock.RLock()
	defer func() {
		DLock.RUnlock()
	}()

}

func RandomizeDirectory() {
	go func() {
		for i := 0; i == i; i++ {
			fmt.Println(i)
			fakeaddr := fmt.Sprintf("host%d:20000", i+1024)
			AddNode(fakeaddr, "wtf", 200, true)
			time.Sleep(time.Minute / time.Duration(10*(rand.Int()%10+2)))
		}
	}()
	for i := 0; i == i; i++ {
		fmt.Println(i)
		if len(KDirectory) > 0 {
			DeleteNode(rand.Int() % len(KDirectory))
		}
		time.Sleep(time.Minute / time.Duration(10*(rand.Int()%10+1)))
	}
}

func ReadKeys() {
	f, e := os.Open("private.key")
	defer f.Close()
	if e != nil {
		panic("cannot of openings")
	}
	r := packet.NewReader(f)
	OurRSAKey, _ = openpgp.ReadEntity(r)
}

func PublishKeys() {
	f, e := os.Create("public.key")
	defer f.Close()
	if e != nil {
		panic("cannot of savings")
	}
	OurRSAKey.Serialize(f)
}

func main() {
	ReadKeys()
	PublishKeys()
	go lifepatrol()
	http.HandleFunc("/read", ReadDirectoryHandler)
	http.HandleFunc("/longpoll", LPDirectoryHandler)
	http.HandleFunc("/rformat", RFormatDirectoryHandler)
	http.HandleFunc("/upload", UploadInfoHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err.Error())
	}
}
