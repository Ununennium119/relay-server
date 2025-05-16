package models

type Lobby struct {
	ID        string
	Clients   map[string]*Client // key: client ID
	CreatedAt int64              // Unix timestamp
}
