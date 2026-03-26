package chat_repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/chat/chat_models"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
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
		return nil, fmt.Errorf("inserting second member error: %w", err)
	}

	return chat, nil
}

func (repo *PostgresRepo) GetDirectByIdRepo(ctx context.Context, MyId uuid.UUID, CompanionId uuid.UUID) (*chat_models.Direct, error) {
	var direct chat_models.Direct
	stmt, err := repo.db.PrepareContext(ctx, `SELECT c.id, c.creation_time,
        m1.user_id as first_user_id, m1.joined_time as first_joined,
        m2.user_id as second_user_id, m2.joined_time as second_joined
    FROM chats c
    JOIN chat_members m1 ON m1.chat_id = c.id AND m1.user_id = $1
    JOIN chat_members m2 ON m2.chat_id = c.id AND m2.user_id = $2
    LIMIT 1`)
	if err != nil {
		return nil, fmt.Errorf("cannot find such direct: %w", err)
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, MyId, CompanionId).Scan(
		&direct.Id,
		&direct.CreationTime,
		&direct.FirstMember.UserId,
		&direct.FirstMember.JoinedTime,
		&direct.SecondMember.UserId,
		&direct.SecondMember.JoinedTime,
	)
	if err == sql.ErrNoRows {
		fmt.Println("Chat not found ")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find direct error: %w", err)
	}
	direct.FirstMember.ChatId = direct.Id
	direct.SecondMember.ChatId = direct.Id

	return &direct, nil
}

func (repo *PostgresRepo) GetListOfDirectsListByIDRepo(ctx context.Context, userId uuid.UUID) ([]*chat_models.Direct, error) {
	stmt, err := repo.db.PrepareContext(ctx, `
        SELECT c.id, c.creation_time,
            m1.user_id as first_user_id, m1.joined_time as first_joined,
            m2.user_id as second_user_id, m2.joined_time as second_joined
        FROM chats c
        JOIN chat_members m1 ON m1.chat_id = c.id AND m1.user_id = $1
        JOIN chat_members m2 ON m2.chat_id = c.id AND m2.user_id != $1
    `)
	if err != nil {
		return nil, fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("query directs error: %w", err)
	}
	defer rows.Close()

	var directs []*chat_models.Direct
	for rows.Next() {
		var direct chat_models.Direct
		err := rows.Scan(
			&direct.Id,
			&direct.CreationTime,
			&direct.FirstMember.UserId,
			&direct.FirstMember.JoinedTime,
			&direct.SecondMember.UserId,
			&direct.SecondMember.JoinedTime,
		)
		if err != nil {
			return nil, fmt.Errorf("scan direct error: %w", err)
		}
		direct.FirstMember.ChatId = direct.Id
		direct.SecondMember.ChatId = direct.Id
		directs = append(directs, &direct)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return directs, nil
}

func (repo *PostgresRepo) SendMessageRepo(ctx context.Context, message *chat_models.Message) (*chat_models.Message, error) {
	stmt, err := repo.db.PrepareContext(ctx, `
        INSERT INTO messages (id, chat_id, sender_id, content, status, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `)
	if err != nil {
		return nil, fmt.Errorf("prepare statement error : %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, message.ID, message.ChatID, message.SenderID, message.Content, message.Status, message.CreatedAT)
	if err != nil {
		return nil, fmt.Errorf("insert message error: %w", err)
	}

	return message, nil
}

func (repo *PostgresRepo) GetMessagesByChatIdRepo(ctx context.Context, chatId uuid.UUID, before time.Time, limit int) ([]*chat_models.Message, error) {
	stmt, err := repo.db.PrepareContext(ctx, `
		SELECT id, chat_id, sender_id, content, status, created_at
		FROM messages
		WHERE chat_id = $1 AND created_at < $2
		ORDER BY created_at DESC
		LIMIT $3
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare statement error: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, chatId, before, limit)
	if err != nil {
		return nil, fmt.Errorf("query messages error: %w", err)
	}
	defer rows.Close()

	var messages []*chat_models.Message = make([]*chat_models.Message, 0)
	for rows.Next() {
		var message chat_models.Message
		err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.Content,
			&message.Status,
			&message.CreatedAT,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message error: %w", err)
		}
		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}

// SearchMesageRepo implements [MessageRepo].
func (repo *PostgresRepo) SearchMesageRepo(ctx context.Context, chat_id uuid.UUID, content string) ([]*chat_models.Message, error) {
	stmt, err := repo.db.PrepareContext(ctx, ` SELECT id, chat_id, sender_id, content, status, created_at
        FROM messages
        WHERE chat_id = $1 AND content ILIKE $2 
        ORDER BY created_at DESC`) // от нового к старому;  ilike для того чтоб не надо было учитвывть регистр ("Привет" == "привет")
	if err != nil {
		return nil, fmt.Errorf("statement error: %w", err)
	}

	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, chat_id, "%"+content+"%")
	//content так для того чтоб не обращать внимание на знаки ("Привет!" == "Привет")
	if err != nil {
		return nil, fmt.Errorf("error of searching : %w ", err)
	}
	defer rows.Close()

	messages := make([]*chat_models.Message, 0)
	for rows.Next() {
		var message chat_models.Message
		err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.CreatedAT,
			&message.Content,
			&message.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("error of scan message : %w", err)
		}
		messages = append(messages, &message)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error returning message : %w", err)
	}

	return messages, nil
	}


	
	func (repo *PostgresRepo) UpdateMessageStatusRepo(ctx context.Context, message_id uuid.UUID, status chat_models.MessageStatus) error {
		stmt, err := repo.db.PrepareContext(ctx, `UPDATE messages SET status = $1 WHERE id = $2 `)
		if err != nil {
			return  fmt.Errorf("error prepare statement : %w", err)
		}
		defer stmt.Close()
		_, err = stmt.ExecContext(ctx,status,message_id) 
		if err != nil {
			return fmt.Errorf("error of updating message status : %w", err)
		}
		return nil 
	}

func (repo *PostgresRepo) DeleteMessageRepo(ctx context.Context, message_id uuid.UUID) error {
	stmt, err := repo.db.PrepareContext(ctx, `DELETE FROM messages WHERE id = $1`)
	if err != nil {
		return fmt.Errorf("prepare statement error : %w", err)
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, message_id)
	if err != nil {
		return fmt.Errorf("delete message error : %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error : %w", err)
	}
	if n == 0 {
		return fmt.Errorf("message not found")
	}
	return nil
}
