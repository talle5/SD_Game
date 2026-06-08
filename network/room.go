package network

import (
	"encoding/json"
	"fmt"
	"net"
	"raylib-game/game"
	"time"
)

const (
	topicSalas = "game/salas"
	topicSinal = "game/signal"
	sala       = "sala1"
)

var Peers = map[string]*net.UDPAddr{}
var Players = map[string]*game.Player{}

type Room struct {
	udpConn *net.UDPConn
	sala    string
	OnJoin  func(*net.UDPAddr) // callback: peer conectou
	OnLeave func(*net.UDPAddr) // callback: peer saiu
}

// NewRoom cria e inicializa uma nova estrutura Room, configurando os callbacks e canais.
func NewRoom(nomeSala string, udpConn *net.UDPConn, onJoin, onLeave func(*net.UDPAddr)) *Room {
	return &Room{
		sala:    nomeSala,
		udpConn: udpConn,
		OnJoin:  onJoin,
		OnLeave: onLeave,
	}
}

// Connect estabelece a conexão com o broker MQTT, se inscreve no tópico de sinalização
// e verifica se a sala já possui um host. Se não possuir, o usuário se anuncia como host;
// caso contrário, tenta entrar na sala existente.
func (r *Room) Connect() error {

	if _, err := MQTTClient(); err != nil {
		return err
	}
	fmt.Println("[MQTT] Conectado ao broker")

	r.subscreverSinal()

	host := buscarHost()
	if host == nil {
		r.anunciar()
	} else {
		r.entrar(host)
	}

	return nil
}

// anunciar publica as informações do próprio peer no tópico de salas de forma retida,
// indicando que ele é o host aguardando conexões de outros peers.
func (r *Room) anunciar() {
	peer := r.meuPeer()
	payload, _ := json.Marshal(peer)
	PublishMQTT(topicSalas+"/"+r.sala, payload, true)
	fmt.Printf("[Room] Sala '%s' criada — aguardando peers...\n", r.sala)
}

// entrar publica as informações locais no tópico de sinalização para que o host
// e os demais participantes saibam de sua entrada, além de iniciar o processo de hole punching.
func (r *Room) entrar(host *Peer) {
	fmt.Printf("[Room] Entrando na sala via %s:%d\n", host.Ip, host.Port)
	eu := r.meuPeer()
	payload, _ := json.Marshal(eu)
	PublishMQTT(topicSinal+"/"+r.sala, payload, false)
	addr, ok := punch(host.Ip, host.Port, r.udpConn)
	if ok {
		Peers[host.Id] = addr
		r.OnJoin(addr)
	}

}

// buscarHost tenta encontrar um host existente para a sala, aguardando até 1 segundo
// por uma mensagem retida no tópico de salas. Se encontrar, retorna o Peer do host; caso contrário, retorna nil.
func buscarHost() *Peer {
	topic := topicSalas + "/" + sala
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

// subscreverSinal inscreve o peer no tópico de sinalização MQTT para receber notificações
// sempre que um novo peer tentar entrar na sala. Inicia automaticamente o hole punching com o novo peer.
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
        
        addr, ok := resolveremoto(peer, eu, r.udpConn)
        if ok {
            Peers[peer.Id] = addr
            r.OnJoin(addr)
        }
    })
}

func resolveremoto(peer Peer, eu Peer, udpConn *net.UDPConn) (*net.UDPAddr, bool) {
	if peer.Ip == eu.Ip && peer.LocalIp == eu.LocalIp {
		remote, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", peer.Port))
		return remote, true
	} else if peer.Ip == eu.Ip {
		remote, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", peer.LocalIp, peer.Port))
		return remote, true
	} else {
		fmt.Printf("Host remoto encontrado — estabelecendo comunicação direta\n")
		return punch(peer.Ip, peer.Port, udpConn)
	}
}

// meuPeer constrói e retorna a estrutura Peer com as informações de IP público,
// IP local e a porta utilizada pela conexão UDP deste cliente.
func (r *Room) meuPeer() Peer {
	port := r.udpConn.LocalAddr().(*net.UDPAddr).Port
    // Passando o clientID gerado no mqttclient.go para garantir unicidade
	return Peer{Id: clientID, Ip: publicIp, LocalIp: localIp, Port: port}
}

func Broadcast(msg []byte, udpConn *net.UDPConn) {
	fmt.Printf("[Broadcast] Enviando mensagem: %s\n", string(msg))
	for _, addr := range Peers {
		udpConn.WriteToUDP([]byte(msg), addr)
	}
}
