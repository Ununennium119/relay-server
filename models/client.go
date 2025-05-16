package models

import "net"

type Client struct {
	ID       string
	Addr     *net.UDPAddr
	LastSeen int64 // Unix timestamp
	LobbyID  string
}
