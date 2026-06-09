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
	if len(message) != 4 {
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

	xVal, errX := strconv.ParseInt(message[1], 10, 32)
	yVal, errY := strconv.ParseInt(message[2], 10, 32)
	dVal, errD := strconv.ParseInt(message[3], 10, 32)
	if errX != nil || errY != nil || errD != nil {
		return	
	}
	if dVal < 0 || dVal > 4 {
		return
	}

	plr.X = int32(xVal)
	plr.Y = int32(yVal)
	plr.Direction = game.Direction(dVal)
}

func exit(addr *net.UDPAddr, message []string) {}
