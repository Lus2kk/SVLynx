package user_repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	UserExistsByTgID(ctx context.Context, telegramID int64) (bool, error)

	UsernameExists(ctx context.Context, username string) (bool, error)
	NicknameExists(ctx context.Context, nickname string) (bool, error)

	SaveUserTelegram(ctx context.Context, telegramID int64, username, firstName, PhotoURL string) error
	SaveUserEmail(ctx context.Context, email, nickname, name, status, avatar_color string) error

	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByTgID(tctx context.Context, elegramID int64) (*User, error)
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
	SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)

	return exists, err
}

func (r *Repository) UserExistsByTgID(ctx context.Context, TelegramID int64) (bool, error){
	var exists bool

	err := r.db.QueryRow(ctx, `
	SELECT EXISTS(SELECT 1 FROM users WHERE telegram_id = $1)`,TelegramID).Scan(&exists)

	if err != nil{
		return false, err
	}

	return exists, nil
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
	SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) NicknameExists(ctx context.Context, nickname string) (bool, error){
	var exists bool

	err := r.db.QueryRow(ctx, `
	SELECT EXISTS(SELECT 1 FROM users WHERE nickname = $1)`, nickname).Scan(&exists)

	if err != nil{
		return false, err
	}
	
	return exists, nil
}

func (r *Repository) SaveUserEmail(ctx context.Context, email, nickname, name, status, avatarColor string) error {
	_, err := r.db.Exec(ctx, `
	INSERT INTO users (email, nickname, name, status, avatar_color) 
	VALUES ($1, $2, $3, $4, $5)`, email, nickname, name, status, avatarColor)

	return err

}

func (r * Repository) SaveUserTelegram(ctx context.Context, telegramID int64, username, firstName, PhotoURL string) error{
	_, err := r.db.Exec(ctx, `
	INSERT INTO users (telegram_id, username, first_name, photo_url)
	VALUES ($1, $2, $3, $4)`, telegramID, username, firstName, PhotoURL)

	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}

	err := r.db.QueryRow(ctx, `
	SELECT id, email, nickname, name, avatar_color
	FROM users WHERE email=$1`, email).Scan(
		&u.ID,
		&u.Email,  
		&u.Nickname, 
		&u.Name, 
		&u.AvatarColor,
	)

	return u, err
}

func (r *Repository) GetUserByTgID(ctx context.Context, telegramID int64) (*User, error){
	u:= &User{}

	err := r.db.QueryRow(ctx, `
	SELECT id, telegram_id, username, first_name, photo_url
	FROM users WHERE telegram_id = $1`, telegramID).Scan(
		&u.ID, 
		&u.TelegramID,
		&u.Username,
		&u.FirstName,
		&u.PhotoURL,
	)

	return u, err
}