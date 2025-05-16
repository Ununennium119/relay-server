package models

const (
	AuthMagicByte = 0xAE // Magic byte for auth packets
)

type AuthPacket struct {
	Token string `json:"token"`
}

type ClientRole int

const (
	RoleMember ClientRole = iota
	RoleOwner
)
