package repository

import (
	"context"
	"errors"
)

func (r *Repository) GetUserAESKey(ctx context.Context, userID string) (string, error) {
	query := `SELECT aes_key FROM lobby_user WHERE id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var aesKey string
		if err := rows.Scan(&aesKey); err != nil {
			return "", err
		}
		return aesKey, nil
	}

	return "", errors.New("user not found")
}
