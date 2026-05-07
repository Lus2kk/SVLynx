package user_repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	UserExistsByTgID(ctx context.Context, telegramID int64) (bool, error)

	UsernameExists(ctx context.Context, username string) (bool, error)
	NicknameExists(ctx context.Context, nickname string) (bool, error)

	SaveUserTelegram(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) error
	SaveUserEmail(ctx context.Context, email string, nickname *string, name, status, avatarColor string) error

	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByTgID(ctx context.Context, telegramID int64) (*User, error)
	GetUserByUserID(ctx context.Context, userID string) (*User, error)

	UpdateUserProfile(ctx context.Context, id, nickname, name, status, avatarColor string) error
	UpdateTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) error

	SearchUsers(ctx context.Context, currentUserID string, query string, limit int) ([]*User, error)
	GetUserLastSeen(ctx context.Context, userID string) (time.Time, error)
	UpdateUserLastSeen(ctx context.Context, userID string) error
	GetUserStatus(ctx context.Context, userID string) (isOnline bool, lastSeen time.Time, err error)
	SetUserOnlineStatus(ctx context.Context, userID string, isOnline bool) error
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`, email).Scan(&exists)

	return exists, err
}

func (r *Repository) UserExistsByTgID(ctx context.Context, telegramID int64) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE telegram_id = $1)
	`, telegramID).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)
	`, username).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repository) NicknameExists(ctx context.Context, nickname string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE nickname = $1)
	`, nickname).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repository) SaveUserEmail(ctx context.Context, email string, nickname *string, name, status, avatarColor string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (email, nickname, name, status, avatar_color)
		VALUES ($1, $2, $3, $4, $5)
	`, email, nickname, name, status, avatarColor)

	return err
}

func (r *Repository) SaveUserTelegram(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (telegram_id, username, first_name, last_name, photo_url)
		VALUES ($1, $2, $3, $4, $5)
	`, telegramID, username, firstName, lastName, photoURL)

	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(telegram_id, 0),
			COALESCE(username, ''),
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(photo_url, ''),
			COALESCE(email, ''),
			COALESCE(nickname, ''),
			COALESCE(name, ''),
			COALESCE(avatar_color, ''),
			COALESCE(status, ''),
			profile_completed
		FROM users
		WHERE email = $1
	`, email).Scan(
		&u.ID,
		&u.TelegramID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.PhotoURL,
		&u.Email,
		&u.Nickname,
		&u.Name,
		&u.AvatarColor,
		&u.Status,
		&u.ProfileCompleted,
	)

	return u, err
}

func (r *Repository) GetUserByTgID(ctx context.Context, telegramID int64) (*User, error) {
	u := &User{}

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(telegram_id, 0),
			COALESCE(username, ''),
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(photo_url, ''),
			COALESCE(email, ''),
			COALESCE(nickname, ''),
			COALESCE(name, ''),
			COALESCE(avatar_color, ''),
			COALESCE(status, ''),
			profile_completed
		FROM users
		WHERE telegram_id = $1
	`, telegramID).Scan(
		&u.ID,
		&u.TelegramID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.PhotoURL,
		&u.Email,
		&u.Nickname,
		&u.Name,
		&u.AvatarColor,
		&u.Status,
		&u.ProfileCompleted,
	)

	return u, err
}

func (r *Repository) GetUserByUserID(ctx context.Context, userID string) (*User, error) {
	u := &User{}

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(telegram_id, 0),
			COALESCE(username, ''),
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(photo_url, ''),
			COALESCE(email, ''),
			COALESCE(nickname, ''),
			COALESCE(name, ''),
			COALESCE(avatar_color, ''),
			COALESCE(status, ''),
			profile_completed
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&u.ID,
		&u.TelegramID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.PhotoURL,
		&u.Email,
		&u.Nickname,
		&u.Name,
		&u.AvatarColor,
		&u.Status,
		&u.ProfileCompleted,
	)

	return u, err
}

func (r *Repository) UpdateUserProfile(ctx context.Context, id, nickname, name, status, avatarColor string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET nickname = $2,
			name = $3,
			status = $4,
			avatar_color = $5,
			profile_completed = TRUE
		WHERE id = $1
	`, id, nickname, name, status, avatarColor)

	return err
}

func (r *Repository) UpdateTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET username = $2,
			first_name = $3,
			last_name = $4,
			photo_url = $5,
			profile_completed = TRUE
		WHERE telegram_id = $1
	`, telegramID, username, firstName, lastName, photoURL)

	return err
}

func (r *Repository) SearchUsers(
    ctx context.Context,
    currentUserID string,
    query string,
    limit int,
) ([]*User, error) {
    rows, err := r.db.Query(ctx, `
        SELECT
            id,
            COALESCE(telegram_id, 0),
            COALESCE(username, ''),
            COALESCE(first_name, ''),
            COALESCE(last_name, ''),
            COALESCE(photo_url, ''),
            COALESCE(email, ''),
            COALESCE(nickname, ''),
            COALESCE(name, ''),
            COALESCE(avatar_color, ''),
            COALESCE(status, ''),
            profile_completed
        FROM users
        WHERE id != $1
          AND (
            nickname ILIKE '%' || $2 || '%'
            OR username ILIKE '%' || $2 || '%'
            OR name ILIKE '%' || $2 || '%'
          )
        ORDER BY
          CASE
            WHEN nickname ILIKE $2 || '%' THEN 0
            WHEN username ILIKE $2 || '%' THEN 1
            WHEN name ILIKE $2 || '%' THEN 2
            ELSE 3
          END,
          nickname ASC
        LIMIT $3
    `, currentUserID, query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        u := &User{}
        if err := rows.Scan(
            &u.ID,
            &u.TelegramID,
            &u.Username,
            &u.FirstName,
            &u.LastName,
            &u.PhotoURL,
            &u.Email,
            &u.Nickname,
            &u.Name,
            &u.AvatarColor,
            &u.Status,
            &u.ProfileCompleted,
        ); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}

func (r *Repository) GetUserLastSeen(ctx context.Context, userID string) (time.Time, error) {
    var lastSeen *time.Time 

    err := r.db.QueryRow(ctx, `
        SELECT last_seen FROM users WHERE id = $1
    `, userID).Scan(&lastSeen)

    if err != nil {
        return time.Time{}, err
    }

    if lastSeen == nil {
        return time.Time{}, nil 
    }

    return *lastSeen, nil
}

func (r *Repository) UpdateUserLastSeen(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `
        UPDATE users SET last_seen = NOW() WHERE id = $1
    `, userID)
	return err
}

func (r *Repository) GetUserStatus(ctx context.Context, userID string) (bool, time.Time, error) {
	var isOnline *bool
	var lastSeen *time.Time

	err := r.db.QueryRow(ctx, `SELECT is_online, last_seen FROM users WHERE id = $1`, userID).Scan(&isOnline, &lastSeen)
	if err != nil {
		return false, time.Time{}, err
	}

	online := false
	if isOnline != nil {
		online = *isOnline
	}

	var seen time.Time
	if lastSeen != nil {
		seen = *lastSeen
	}

	return online, seen, nil
}

func (r *Repository) SetUserOnlineStatus(ctx context.Context, userID string, isOnline bool) error {
	if isOnline {
		_, err := r.db.Exec(ctx, `
			UPDATE users SET is_online = TRUE WHERE id = $1
		`, userID)
		return err
	}
	_, err := r.db.Exec(ctx, `
		UPDATE users SET is_online = FALSE, last_seen = NOW() WHERE id = $1
	`, userID)
	return err
}