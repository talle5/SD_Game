package network

import (
	"fmt"
	"net"
	"time"
)

func punch(ip string, port int, udpConn *net.UDPConn) bool {
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	ping := []byte("PUNCH")
	buf := make([]byte, 32)

	fmt.Printf("[Punch] Iniciando com %s:%d\n", ip, port)

	// envia PUNCH em goroutine separada
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				udpConn.WriteToUDP(ping, addr)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	// aguarda receber PUNCH do outro lado
	udpConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer udpConn.SetReadDeadline(time.Time{})
	defer close(done)

	for {
		n, raddr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("[Punch] ❌ Timeout — sem resposta")
			return false
		}

		msg := string(buf[:n])
		if msg == "PUNCH" || msg == "PUNCH_OK" {
			udpConn.WriteToUDP([]byte("PUNCH_OK"), raddr)
			fmt.Printf("[Punch] ✅ Buraco aberto com %s\n", raddr)
			return true
		}
	}
}