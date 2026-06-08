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
			// Envia o addr para o canal do punch.go destravar
			select {
			case punchChan <- addr:
			default:
			}
		case "move":
			fmt.Print(texto)
			move(addr, mensagem[1:msize])
		case "exit":
			exit(addr, mensagem[1:msize])
		default:
			if texto != "" {
				fmt.Printf("[UDP] Mensagem desconhecida: %s\n", texto)
			}
		}
	}
}

func move(addr *net.UDPAddr, message []string) {
    // Busca o jogador convertendo o endereço UDP para texto ("IP:Porta")
	plr := Players[addr.String()] 
	if plr == nil {
		return
	}
	
	plr.X = atoi32(message[0])
	plr.Y = atoi32(message[1])
	plr.Direction = game.Direction(atoi32(message[2]))
}

func exit(addr *net.UDPAddr, message []string) {}

func atoi32(s string) int32 {
	i, _ := strconv.Atoi(s)
	return int32(i)
}
