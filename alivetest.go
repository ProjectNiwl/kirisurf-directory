// alivetest.go
package main

import (
	"fmt"
	"net"
	"time"
)

func IsAlive(thing *KNode) bool {
	remaddr := thing.Address
	conn, err := net.DialTimeout("tcp", remaddr, time.Second*10)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
