package udp

import (
	"bytes"
	"errors"
	"log"
	"net"
	"sync/atomic"
)

const (
	defaultMTU = 1024
)

//UDPConfig ...
type UDPConfig struct {
	Server  Serverconf
	Backend []Backendconf
}

//Serverconf ...
type Serverconf struct {
	Name string
	Addr string
}

//Backendconf ...
type Backendconf struct {
	Name     string
	Location string
	MTU      int
}

// UDP this is a whole struct to represent the relay + backend
type UDP struct {
	addr string
	name string

	closing int64
	l       *net.UDPConn

	backends []*udpBackend
}

type udpBackend struct {
	u    *UDP
	s    *net.UDPConn
	name string
	addr *net.UDPAddr
	mtu  int
}

//InitUDPServer will load the conf and start the the socket
func InitUDPServer(config UDPConfig) (*UDP, error) {
	u := new(UDP)

	u.name = config.Server.Name
	u.addr = config.Server.Addr

	l, err := net.ListenPacket("udp", u.addr)
	if err != nil {
		return nil, err
	}

	ul, ok := l.(*net.UDPConn)
	if !ok {
		return nil, errors.New("problem listening for UDP")
	}

	u.l = ul

	for i := range config.Backend {
		cfg := &config.Backend[i]

		if cfg.MTU == 0 {
			cfg.MTU = defaultMTU
		}

		addr, err := net.ResolveUDPAddr("udp", cfg.Location)
		if err != nil {
			return nil, err
		}

		s, err := net.ListenUDP("udp", nil)
		if err != nil {
			return nil, err
		}
		u.backends = append(u.backends, &udpBackend{u, s, cfg.Name, addr, cfg.MTU})
	}

	return u, nil
}

func (u *UDP) processPacket(queue chan []byte) {

	var err error

	for p := range queue {

		for _, b := range u.backends {
			_, err = b.s.WriteToUDP(p, b.addr)
			if err != nil {
				log.Printf("Error writing UDP to backend %s: %v", b.name, err)
			}
		}
	}
}

//Start ...
func (u *UDP) Start() error {

	buf := make([]byte, 2<<16)
	queue := make(chan []byte)

	go u.processPacket(queue)

	for {
		n, _, err := u.l.ReadFromUDP(buf[:])
		if err != nil {
			log.Println("error: Failed to receive packet", err)
			continue
		}

		var queuebuffer bytes.Buffer
		queuebuffer.Write(buf[:n])
		queue <- queuebuffer.Bytes()

	}

}

//Stop ...
func (u *UDP) Stop() error {
	atomic.StoreInt64(&u.closing, 1)
	return u.l.Close()
}
