package network

import (
	"fmt"
	"net"
	"time"
)

var punchChan = make(chan *net.UDPAddr, 1)

func punch(ip string, port int, udpConn *net.UDPConn) (*net.UDPAddr, bool) {
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	ping := []byte("PUNCH")

	go func() {
		for {
			udpConn.WriteToUDP(ping, addr)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {
	case raddr := <-punchChan:
		return raddr, true
	case <-time.After(5 * time.Second):
		return nil, false
	}
}
