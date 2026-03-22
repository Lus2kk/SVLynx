package chat_repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/svlynx/messenger/internal/chat/chat_models"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo (db *sql.DB) *PostgresRepo {
    return &PostgresRepo{db: db}
}

func (repo *PostgresRepo) CreateNewDirectRepo(ctx context.Context, chat *chat_models.Direct) (*chat_models.Direct, error) {
    stmt, err := repo.db.PrepareContext(ctx, `
        INSERT INTO chats (id, creation_time)
        VALUES ($1, $2)`)
    if err != nil {
        return nil, fmt.Errorf("prepare statement: %w", err)
    }
    defer stmt.Close()

    _, err = stmt.ExecContext(ctx, chat.Id, chat.CreationTime)
    if err != nil {
        return nil, fmt.Errorf("creating chat error: %w", err)
    }
   memberStmt, err := repo.db.PrepareContext(ctx, `
        INSERT INTO chat_members (chat_id, user_id, joined_time)
        VALUES ($1, $2, $3)`)
    if err != nil {
        return nil, fmt.Errorf("prepare member statement: %w", err)
    }
    defer memberStmt.Close()

    _, err = memberStmt.ExecContext(ctx, chat.FirstMember.ChatId, chat.FirstMember.UserId, chat.FirstMember.JoinedTime)
    if err != nil {
        return nil, fmt.Errorf("inserting first member error: %w", err)
    }

    _, err = memberStmt.ExecContext(ctx, chat.SecondMember.ChatId, chat.SecondMember.UserId, chat.SecondMember.JoinedTime)
    if err != nil {
       return nil, fmt.Errorf("inserting second member error: %w",err)
    }

    return chat, nil
}
