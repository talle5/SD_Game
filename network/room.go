package network

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	topicSalas = "game/salas"
	topicSinal = "game/signal"
)

type Room struct {
	udpConn  *net.UDPConn
	remote   *net.UDPAddr
	sala     string
	vizinhos []Peer
	NovoPeer chan Peer
}

func NewRoom(nomeSala string, udpConn *net.UDPConn) *Room {
	return &Room{
		sala:     nomeSala,
		udpConn:  udpConn,
		NovoPeer: make(chan Peer),
	}
}

func (r *Room) Connect() error {

	if _, err := MQTTClient(); err != nil {
		return err
	}
	fmt.Println("[MQTT] Conectado ao broker")

	r.subscreverSinal()

	host := r.buscarHost()
	if host == nil {
		r.anunciar()
	} else {
		r.entrar(host)
	}

	return nil
}

func (r *Room) anunciar() {
	peer := r.meuPeer()
	payload, _ := json.Marshal(peer)
	PublishMQTT(topicSalas+"/"+r.sala, payload)
	fmt.Printf("[Room] Sala '%s' criada — aguardando peers...\n", r.sala)
}

func (r *Room) entrar(host *Peer) {
	fmt.Printf("[Room] Entrando na sala via %s:%d\n", host.Ip, host.Port)

	eu := r.meuPeer()
	payload, _ := json.Marshal(eu)
	PublishMQTTRetained(topicSinal+"/"+r.sala, payload, false)

	if host.Ip == eu.Ip && host.LocalIp == eu.LocalIp {
		fmt.Println("[Room] Mesmo computador → loopback")
		r.remote, _ = net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", host.Port))
	} else if host.Ip == eu.Ip {
		fmt.Printf("[Room] Mesma rede → IP local %s:%d\n", host.LocalIp, host.Port)
		r.remote, _ = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host.LocalIp, host.Port))
	} else {
		punch(host.Ip, host.Port, r.udpConn)
	}
}

func (r *Room) buscarHost() *Peer {
	topic := topicSalas + "/" + r.sala
	result := make(chan *Peer, 1)

	SubscribeMQTT(topic, func(payload []byte) {
		var peer Peer
		if err := json.Unmarshal(payload, &peer); err == nil {
			result <- &peer
		} else {
			result <- nil
		}
	})

	select {
	case peer := <-result:
		return peer
	case <-time.After(1 * time.Second):
		return nil
	}
}

func (r *Room) subscreverSinal() {
	topic := topicSinal + "/" + r.sala
	SubscribeMQTT(topic, func(payload []byte) {
		var peer Peer
		if err := json.Unmarshal(payload, &peer); err != nil {
			return
		}

		eu := r.meuPeer()
		if peer.Ip == eu.Ip && peer.LocalIp == eu.LocalIp && peer.Port == eu.Port {
			return
		}

		// define rota UDP
		if peer.Ip == eu.Ip && peer.LocalIp == eu.LocalIp {
			r.remote, _ = net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", peer.Port))
		} else if peer.Ip == eu.Ip {
			r.remote, _ = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", peer.LocalIp, peer.Port))
		} else {
			go func() {
				ok := punch(peer.Ip, peer.Port, r.udpConn)
				if ok {
					r.NovoPeer <- peer
				}
			}()
		}
		go func() {
			ok := punch(peer.Ip, peer.Port, r.udpConn)
			if ok {
				r.NovoPeer <- peer
			}
		}()
	})
}

func (r *Room) meuPeer() Peer {
	if publicIp == "" {
		publicIp = obterIpPublico()
	}
	if localIp == "" {
		localIp = obterIpLocal()
	}
	port := r.udpConn.LocalAddr().(*net.UDPAddr).Port
	return Peer{Ip: publicIp, LocalIp: localIp, Port: port}
}
