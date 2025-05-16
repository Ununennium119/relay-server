package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

func (r *Repository) GetLobbyMembers(ctx context.Context, lobbyID string) ([]string, error) {
	query := `SELECT user_id FROM lobby_member WHERE lobby_id = $1`
	rows, err := r.db.QueryContext(ctx, query, lobbyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		members = append(members, userID)
	}

	return members, nil
}

func (r *Repository) GetUserLobby(ctx context.Context, userID uuid.UUID) (string, bool, error) {
	query := `SELECT lobby_id, is_owner FROM lobby_member WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return "", false, err
	}
	defer rows.Close()

	for rows.Next() {
		var lobbyID string
		var isOwner bool
		if err := rows.Scan(&lobbyID, &isOwner); err != nil {
			return "", false, err
		}
		return lobbyID, isOwner, nil
	}

	return "", false, errors.New("lobby not found")
}

func (r *Repository) IsUserInLobby(ctx context.Context, userID, lobbyID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM lobby_member WHERE user_id = $1 AND lobby_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, lobbyID).Scan(&exists)
	return exists, err
}
