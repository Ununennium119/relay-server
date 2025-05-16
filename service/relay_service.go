package service

import (
	"Relay/models"
	"Relay/repository"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net"
	"sync"
)

type RelayService struct {
	userRepo           *repository.Repository
	lobbyRepo          *repository.Repository
	authService        *AuthService
	ownerMapping       map[uuid.UUID]uuid.UUID              // lobbyID -> ownerUserID
	peerMapping        map[uuid.UUID]map[uuid.UUID]struct{} // userID -> userId
	connections        map[uuid.UUID]*net.UDPAddr
	reverseConnections map[string]uuid.UUID
	mutex              sync.RWMutex
}

func NewRelayService(
	userRepo *repository.Repository,
	lobbyRepo *repository.Repository,
	authService *AuthService,
) *RelayService {
	return &RelayService{
		userRepo:           userRepo,
		lobbyRepo:          lobbyRepo,
		authService:        authService,
		ownerMapping:       make(map[uuid.UUID]uuid.UUID),
		peerMapping:        make(map[uuid.UUID]map[uuid.UUID]struct{}),
		connections:        make(map[uuid.UUID]*net.UDPAddr),
		reverseConnections: make(map[string]uuid.UUID),
	}
}

func (s *RelayService) HandlePacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	if len(data) == 0 {
		return
	}

	// Check for auth packet (first byte is magic byte)
	if data[0] == models.AuthMagicByte {
		s.handleAuthPacket(conn, addr, data[1:])
		return
	}

	// Regular data packet
	s.handleDataPacket(conn, addr, data)
}

func (s *RelayService) handleAuthPacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	if len(data) < 36+12+16 {
		log.Println("Invalid encrypted packet size")
		return
	}

	userID, err := uuid.Parse(string(data[:36]))
	if err != nil {
		return
	}

	base64AESKey, err := s.userRepo.GetUserAESKey(context.Background(), userID.String())
	if err != nil {
		log.Println("Failed to get user:", err)
		return
	}

	aesKey, err := base64.StdEncoding.DecodeString(base64AESKey)
	if err != nil {
		log.Fatalf("Failed to decode base64 key: %v", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		log.Println("Failed to create cipher:", err)
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Println("Failed to create AES-GCM:", err)
		return
	}

	iv := data[36:48]
	ciphertext := data[48:]
	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		log.Println("Failed to decrypt auth packet:", err)
		return
	}

	var auth models.AuthPacket
	if err := json.Unmarshal(plaintext, &auth); err != nil {
		return
	}

	claims, err := s.authService.ValidateToken(auth.Token)
	if err != nil {
		return
	}

	tokenUserID, err := uuid.Parse(claims["sub"].(string))
	if err != nil || tokenUserID != userID {
		return
	}

	// Get lobby info from database
	lobbyIDString, isOwner, err := s.lobbyRepo.GetUserLobby(context.Background(), userID)
	if err != nil {
		return
	}

	lobbyID, err := uuid.Parse(lobbyIDString)
	if err != nil {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store connection
	s.connections[userID] = addr
	s.reverseConnections[addr.IP.String()+":"+string(rune(addr.Port))] = userID

	if isOwner {
		// Store owner mapping
		s.ownerMapping[lobbyID] = userID
	} else {
		// Map this user to their owner
		if ownerID, exists := s.ownerMapping[lobbyID]; exists {
			if s.peerMapping[userID] == nil {
				s.peerMapping[userID] = make(map[uuid.UUID]struct{})
			}
			if s.peerMapping[ownerID] == nil {
				s.peerMapping[ownerID] = make(map[uuid.UUID]struct{})
			}
			s.peerMapping[userID][ownerID] = struct{}{}
			s.peerMapping[ownerID][userID] = struct{}{}
		}
	}

	// Send ACK to client
	ack := map[string]any{
		"ok": true,
	}
	ackBytes, _ := json.Marshal(ack)
	conn.WriteToUDP(append([]byte{models.AuthMagicByte}, ackBytes...), addr)
}

func (s *RelayService) handleDataPacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Find user by address
	var userID = s.reverseConnections[addr.IP.String()+":"+string(rune(addr.Port))]
	if userID == uuid.Nil {
		return
	}

	// Check if this user is mapped to a user
	if peers, exists := s.peerMapping[userID]; exists {
		for peerID := range peers {
			if ownerAddr, ok := s.connections[peerID]; ok {
				// Send to each owner
				conn.WriteToUDP(data, ownerAddr)
			}
		}
	}
}

func (s *RelayService) CleanupOldConnections() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Implementation for cleaning up old connections
	// Would check last seen times and remove stale entries
}
