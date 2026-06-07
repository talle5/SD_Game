package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"raylib-game/game"
)

var (
	peers  = make(map[string]*Peer)
	player = make(map[Peer]*game.Player)
)

func ReceiverLoop(socket *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		n, _, _ := socket.ReadFromUDP(buf)
		texto := string(buf[:n])
		mensagem := strings.Split(texto, " ")
		msize := len(mensagem)
		if msize == 0 {
			continue
		}
		switch mensagem[0] {
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
	peer := peers[id]
	if peer == nil {
		return
	}
	plr := player[*peer]
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
