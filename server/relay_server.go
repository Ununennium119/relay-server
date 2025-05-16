package server

import (
	"Relay/service"
	"net"
)

type RelayServer struct {
	port         int
	relayService *service.RelayService
}

func NewRelayServer(port int, relayService *service.RelayService) *RelayServer {
	return &RelayServer{
		port:         port,
		relayService: relayService,
	}
}

func (s *RelayServer) Start() error {
	addr := net.UDPAddr{
		Port: s.port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, 1500)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		go s.relayService.HandlePacket(conn, addr, buf[:n])
	}
}
