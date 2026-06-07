package network

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
)

const (
	broadcastDiscoveryPort = 47777
	broadcastMessageType   = "raylib-game/peer"
	broadcastInterval      = 1 * time.Second
)

type broadcastMessage struct {
	Type string `json:"type"`
	Room string `json:"room"`
	Peer Peer   `json:"peer"`
}

// BroadcastDiscovery anuncia este peer na rede local e escuta outros peers por UDP broadcast.
type BroadcastDiscovery struct {
	room string
	self Peer

	conn *net.UDPConn

	done  chan struct{}
	peers chan Peer

	mu    sync.Mutex
	seen  map[string]Peer
	once  sync.Once
	start sync.Once
}

// NewBroadcastDiscovery cria uma descoberta LAN para a sala informada.
func NewBroadcastDiscovery(room string, self Peer) *BroadcastDiscovery {
	if self.Id == "" {
		self.Id = newPeerID()
	}
	if self.Name == "" {
		self.Name, _ = os.Hostname()
	}

	return &BroadcastDiscovery{
		room:  room,
		self:  self,
		done:  make(chan struct{}),
		peers: make(chan Peer, 16),
		seen:  make(map[string]Peer),
	}
}

// Start abre a porta de descoberta e inicia as goroutines de escuta e anúncio.
func (d *BroadcastDiscovery) Start() error {
	var err error
	d.start.Do(func() {
		d.conn, err = listenBroadcastUDP(broadcastDiscoveryPort)
		if err != nil {
			return
		}

		go d.listen()
		go d.announceLoop()
		err = d.AnnounceOnce()
	})
	return err
}

// Stop encerra a descoberta e libera a porta UDP usada pelo broadcast.
func (d *BroadcastDiscovery) Stop() {
	d.once.Do(func() {
		close(d.done)
		if d.conn != nil {
			d.conn.Close()
		}
	})
}

// Peers retorna um canal que recebe cada peer novo descoberto na rede local.
func (d *BroadcastDiscovery) Peers() <-chan Peer {
	return d.peers
}

// KnownPeers retorna uma cópia dos peers já descobertos.
func (d *BroadcastDiscovery) KnownPeers() []Peer {
	d.mu.Lock()
	defer d.mu.Unlock()

	peers := make([]Peer, 0, len(d.seen))
	for _, peer := range d.seen {
		peers = append(peers, peer)
	}
	return peers
}

// AnnounceOnce envia um anúncio deste peer para os endereços de broadcast locais.
func (d *BroadcastDiscovery) AnnounceOnce() error {
	if d.conn == nil {
		return errors.New("broadcast discovery is not started")
	}

	payload, err := json.Marshal(broadcastMessage{
		Type: broadcastMessageType,
		Room: d.room,
		Peer: d.self,
	})
	if err != nil {
		return err
	}

	var sendErr error
	for _, addr := range broadcastAddrs(broadcastDiscoveryPort) {
		if _, err := d.conn.WriteToUDP(payload, addr); err != nil {
			sendErr = err
		}
	}
	return sendErr
}

func (d *BroadcastDiscovery) announceLoop() {
	ticker := time.NewTicker(broadcastInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.done:
			return
		case <-ticker.C:
			_ = d.AnnounceOnce()
		}
	}
}

func (d *BroadcastDiscovery) listen() {
	buf := make([]byte, 2048)

	for {
		n, addr, err := d.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-d.done:
				return
			default:
				continue
			}
		}

		var msg broadcastMessage
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			continue
		}
		if msg.Type != broadcastMessageType || msg.Room != d.room {
			continue
		}
		if msg.Peer.Id != "" && msg.Peer.Id == d.self.Id {
			continue
		}

		peer := msg.Peer
		if peer.LocalIp == "" {
			peer.LocalIp = addr.IP.String()
		}
		if peer.Port == 0 {
			peer.Port = addr.Port
		}

		d.addPeer(peer)
	}
}

func (d *BroadcastDiscovery) addPeer(peer Peer) {
	key := peer.Id
	if key == "" {
		key = fmt.Sprintf("%s:%d", peer.LocalIp, peer.Port)
	}

	d.mu.Lock()
	_, exists := d.seen[key]
	d.seen[key] = peer
	d.mu.Unlock()

	if exists {
		return
	}

	select {
	case d.peers <- peer:
	default:
	}
}

func listenBroadcastUDP(port int) (*net.UDPConn, error) {
	config := net.ListenConfig{
		Control: func(_, _ string, c syscall.RawConn) error {
			var controlErr error
			if err := c.Control(func(fd uintptr) {
				controlErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			}); err != nil {
				return err
			}
			return controlErr
		},
	}

	packetConn, err := config.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	conn, ok := packetConn.(*net.UDPConn)
	if !ok {
		packetConn.Close()
		return nil, errors.New("broadcast listener is not UDP")
	}
	return conn, nil
}

func broadcastAddrs(port int) []*net.UDPAddr {
	addrs := []*net.UDPAddr{{IP: net.IPv4bcast, Port: port}}

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		ifaceAddrs, _ := iface.Addrs()
		for _, addr := range ifaceAddrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			mask := ipNet.Mask
			if ip == nil || len(mask) != net.IPv4len {
				continue
			}

			broadcast := net.IPv4(
				ip[0]|^mask[0],
				ip[1]|^mask[1],
				ip[2]|^mask[2],
				ip[3]|^mask[3],
			)
			addrs = append(addrs, &net.UDPAddr{IP: broadcast, Port: port})
		}
	}

	return addrs
}

func newPeerID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err == nil {
		return hex.EncodeToString(buf[:])
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
