package network

import (
	"fmt"
	"net"
	"raylib-game/game"
	"strconv"
	"strings"
)

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
			select {
			case punchChan <- addr:
			default:
			}
		case "move":
			fmt.Println(texto)
			move(mensagem[1:msize])
		case "exit":
			exit(addr, mensagem[1:msize])
		default:
			if texto != "" {
				fmt.Printf("[UDP] Mensagem desconhecida: %s\n", texto)
			}
		}
	}
}

func move(message []string) {
	if len(message) < 4 {
		return
	}

	id := message[0]
	
	peer := Peers[id]
	if peer == nil {
		return
	}
	plr := Players[id] 
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

func exit(addr *net.UDPAddr, message []string) {}

func atoi32(s string) int32 {
	i, _ := strconv.Atoi(s)
	return int32(i)
}
