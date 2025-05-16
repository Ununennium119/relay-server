package models

const (
	AuthMagicByte = 0xAE // Magic byte for auth packets
)

type AuthPacket struct {
	Token string `json:"token"`
}
