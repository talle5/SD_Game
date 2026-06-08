package network

import (
	"fmt"
	"net"
	"raylib-game/game"
	"strconv"
	"strings"
)

var punchDone = make(chan bool, 1)

func ReceiverLoop(socket *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		n, addr, _ := socket.ReadFromUDP(buf)
		texto := string(buf[:n])
		mensagem := strings.Split(texto, " ")
		msize := len(mensagem)
		if msize == 0 {
			continue
		}
		switch mensagem[0] {
		case "PUNCH", "PUNCH_OK":
			socket.WriteToUDP([]byte("PUNCH_OK"), addr)
			punchDone <- true
		case "move":
			move(mensagem[1:msize])
		case "exit":
			exit(mensagem[1:msize])
		default:
			fmt.Printf("[UDP] Mensagem desconhecida: %s\n", texto)
		}
	}
}

func move(message []string) {
	id := message[0]
	peer := Peers[id]
	if peer == nil {
		return
	}
	plr := Players[peer]
	if plr == nil {
		return
	}
	x := message[1]
	y := message[2]
	d := message[3]
	plr.X = atoi32(x)
	plr.Y = atoi32(y)
	plr.Direction = game.Direction(atoi32(d))
}

func exit(message []string) {}

func atoi32(s string) int32 {
	i, _ := strconv.Atoi(s)
	return int32(i)
}
