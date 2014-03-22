// alivetest.go
package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func is_alive(thing KNode) bool {
	remaddr := thing.Address
	conn, err := net.DialTimeout("tcp", remaddr, time.Second*10)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// This is the goroutine that is spawned to patrol the list, removing
// offline nodes
func lifepatrol() {
	for {
		if len(KDirectory) > 0 {
			DLock.RLock()
			idx := rand.Int() % len(KDirectory)
			tgt := KDirectory[idx]
			DLock.RUnlock()
			fmt.Printf("Patrolling life of %s\n", tgt.Address)
			addr := tgt.Address
			if is_alive(tgt) {
				fmt.Printf("%s is alive!\n", addr)
			} else {
				fmt.Printf("%s is dead!\n", addr)
				DeleteNode(idx)
			}

		}
		time.Sleep(time.Second * 5)
	}
}
