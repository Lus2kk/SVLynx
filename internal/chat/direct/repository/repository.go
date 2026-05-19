package chat_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	chat_models "github.com/svlynx/messenger/internal/chat/models"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func replyToJSON(rt *chat_models.ReplyToMessage) []byte {
	if rt == nil {
		return nil
	}
	b, _ := json.Marshal(rt)
	return b
}

func (repo *PostgresRepo) CreateNewDirectRepo(ctx context.Context, chat *chat_models.Direct) (*chat_models.Direct, error) {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO chats (id, creation_time)
		VALUES ($1, $2)`,
		chat.Id, chat.CreationTime)
	if err != nil {
		return nil, fmt.Errorf("creating chat error: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO chat_members (chat_id, user_id, joined_time)
		VALUES ($1, $2, $3)`,
		chat.FirstMember.ChatId, chat.FirstMember.UserId, chat.FirstMember.JoinedTime)
	if err != nil {
		return nil, fmt.Errorf("inserting first member error: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO chat_members (chat_id, user_id, joined_time)
		VALUES ($1, $2, $3)`,
		chat.SecondMember.ChatId, chat.SecondMember.UserId, chat.SecondMember.JoinedTime)
	if err != nil {
		return nil, fmt.Errorf("inserting second member error: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return chat, nil
}

func (repo *PostgresRepo) GetDirectByIdRepo(ctx context.Context, MyId uuid.UUID, CompanionId uuid.UUID) (*chat_models.Direct, error) {
	var direct chat_models.Direct

	err := repo.db.QueryRow(ctx, `
		SELECT c.id, c.creation_time,
			m1.user_id as first_user_id, m1.joined_time as first_joined,
			m2.user_id as second_user_id, m2.joined_time as second_joined
		FROM chats c
		JOIN chat_members m1 ON m1.chat_id = c.id AND m1.user_id = $1
		JOIN chat_members m2 ON m2.chat_id = c.id AND m2.user_id = $2
		LIMIT 1`,
		MyId, CompanionId,
	).Scan(
		&direct.Id,
		&direct.CreationTime,
		&direct.FirstMember.UserId,
		&direct.FirstMember.JoinedTime,
		&direct.SecondMember.UserId,
		&direct.SecondMember.JoinedTime,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find direct error: %w", err)
	}

	direct.FirstMember.ChatId = direct.Id
	direct.SecondMember.ChatId = direct.Id

	return &direct, nil
}

func (repo *PostgresRepo) GetListOfDirectsListByIDRepo(ctx context.Context, userId uuid.UUID) ([]*chat_models.DirectListItem, error) {
	rows, err := repo.db.Query(ctx, `
    SELECT
    c.id,
    c.creation_time,
    m1.user_id AS first_user_id,
    m2.user_id AS second_user_id,
    u.id AS companion_id,
    COALESCE(NULLIF(u.name, ''), NULLIF(TRIM(u.first_name || ' ' || u.last_name), ''), u.username, '') AS companion_name,
    COALESCE(u.nickname, '') AS companion_nickname,
    COALESCE(u.photo_url, '') AS companion_photo_url,
    COALESCE(u.avatar_color, '') AS companion_avatar_color,
    COALESCE(msg.content, '') AS last_message_content,
    msg.created_at AS last_message_at,
    msg.sender_id AS last_message_sender_id,
    COALESCE(msg.status::text, '') AS last_message_status,
    COALESCE(u.is_developer, false) AS companion_is_developer,
    (
        SELECT COUNT(*)
        FROM messages
        WHERE chat_id = c.id
          AND sender_id != $1
          AND status != 'read'
    ) AS unread_count
FROM chats c
JOIN chat_members m1
    ON m1.chat_id = c.id AND m1.user_id = $1
JOIN chat_members m2
    ON m2.chat_id = c.id AND m2.user_id != $1
LEFT JOIN users u
    ON u.id = m2.user_id
LEFT JOIN LATERAL (
    SELECT content, created_at, sender_id, status
    FROM messages
    WHERE chat_id = c.id
    ORDER BY created_at DESC
    LIMIT 1
) msg ON true
ORDER BY COALESCE(msg.created_at, c.creation_time) DESC
`, userId)
	if err != nil {
		return nil, fmt.Errorf("query directs error: %w", err)
	}
	defer rows.Close()

	var directs []*chat_models.DirectListItem

	for rows.Next() {
		var item chat_models.DirectListItem
		var lastMessageAt *time.Time
		var lastMessageSenderId *uuid.UUID
		var lastMessageStatus *string

		err := rows.Scan(
			&item.Id,
			&item.CreationTime,
			&item.FirstUserId,
			&item.SecondUserId,
			&item.CompanionId,
			&item.CompanionName,
			&item.CompanionNickname,
			&item.CompanionPhotoURL,
			&item.CompanionAvatarColor,
			&item.LastMessageContent,
			&lastMessageAt,
			&lastMessageSenderId,
			&lastMessageStatus,
			&item.CompanionIsDeveloper,
			&item.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan direct error: %w", err)
		}
		if lastMessageAt != nil {
			item.LastMessageAt = *lastMessageAt
		}
		if lastMessageSenderId != nil {
			item.LastMessageSenderId = *lastMessageSenderId
		}
		if lastMessageStatus != nil {
			item.LastMessageStatus = *lastMessageStatus
		}

		directs = append(directs, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return directs, nil
}

func (repo *PostgresRepo) SendMessageRepo(ctx context.Context, message *chat_models.Message) (*chat_models.Message, error) {
	_, err := repo.db.Exec(ctx, `
		INSERT INTO messages (id, chat_id, sender_id, content, status, created_at, type, duration, file_name, file_size, reply_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		message.ID, message.ChatID, message.SenderID, message.Content, message.Status, message.CreatedAT, message.Type, message.Duration, message.FileName, message.FileSize, replyToJSON(message.ReplyTo), message.Transcript,
	)
	if err != nil {
		return nil, fmt.Errorf("insert message error: %w", err)
	}

	return message, nil
}

func (repo *PostgresRepo) GetMessagesByChatIdRepo(ctx context.Context, chatId uuid.UUID, before time.Time, limit int) ([]*chat_models.Message, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT id, chat_id, sender_id, content, status, created_at, type, duration, COALESCE(file_name, ''), COALESCE(file_size, 0),  reply_to,  COALESCE(transcript, '')
		FROM messages
		WHERE chat_id = $1 AND created_at < $2
		ORDER BY created_at DESC
		LIMIT $3`,
		chatId, before, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query messages error: %w", err)
	}
	defer rows.Close()

	messages := make([]*chat_models.Message, 0)
	for rows.Next() {
		var message chat_models.Message
		var replyToRaw []byte
		err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.Content,
			&message.Status,
			&message.CreatedAT,
			&message.Type,
			&message.Duration,
			&message.FileName,
			&message.FileSize,
			&replyToRaw,
			&message.Transcript,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message error: %w", err)
		}
		if replyToRaw != nil {
			var rt chat_models.ReplyToMessage
			if err := json.Unmarshal(replyToRaw, &rt); err == nil {
				message.ReplyTo = &rt
			}
		}
		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}

func (repo *PostgresRepo) SearchMesageRepo(ctx context.Context, chat_id uuid.UUID, content string) ([]*chat_models.Message, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT id, chat_id, sender_id, content, status, created_at, type, duration, COALESCE(file_name, ''), COALESCE(file_size, 0), reply_to,  COALESCE(transcript, '')
		FROM messages
		WHERE chat_id = $1 AND content ILIKE $2
		ORDER BY created_at DESC`,
		chat_id, "%"+content+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("search query error: %w", err)
	}
	defer rows.Close()

	messages := make([]*chat_models.Message, 0)
	for rows.Next() {
		var message chat_models.Message
		var replyToRaw []byte
		err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.Content,
			&message.Status,
			&message.CreatedAT,
			&message.Type,
			&message.Duration,
			&message.FileName,
			&message.FileSize,
			&replyToRaw,
			&message.Transcript,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message error: %w", err)
		}
		if replyToRaw != nil {
			var rt chat_models.ReplyToMessage
			if err := json.Unmarshal(replyToRaw, &rt); err == nil {
				message.ReplyTo = &rt
			}
		}
		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}

func (repo *PostgresRepo) UpdateMessageStatusRepo(ctx context.Context, message_id uuid.UUID, status chat_models.MessageStatus) error {
	_, err := repo.db.Exec(ctx, `
		UPDATE messages SET status = $1 WHERE id = $2`,
		status, message_id,
	)
	if err != nil {
		return fmt.Errorf("update message status error: %w", err)
	}
	return nil
}

func (repo *PostgresRepo) DeleteMessageRepo(ctx context.Context, message_id uuid.UUID) error {
	res, err := repo.db.Exec(ctx, `
		DELETE FROM messages WHERE id = $1`,
		message_id,
	)
	if err != nil {
		return fmt.Errorf("delete message error: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("message not found")
	}
	return nil
}

func (repo *PostgresRepo) MarkChatMessagesAsReadRepo(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error {
	_, err := repo.db.Exec(ctx, `
		UPDATE messages
		SET status = 'read'
		WHERE chat_id = $1
		  AND sender_id != $2
		  AND status != 'read'
	`, chatID, userID)
	if err != nil {
		return fmt.Errorf("mark chat messages as read error: %w", err)
	}
	return nil
}

func (repo *PostgresRepo) DeleteDirectRepo(ctx context.Context, chatID uuid.UUID) error {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM messages WHERE chat_id = $1`, chatID)
	if err != nil {
		return fmt.Errorf("delete messages error: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM chat_members WHERE chat_id = $1`, chatID)
	if err != nil {
		return fmt.Errorf("delete chat_members error: %w", err)
	}

	res, err := tx.Exec(ctx, `DELETE FROM chats WHERE id = $1`, chatID)
	if err != nil {
		return fmt.Errorf("delete chat error: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("chat not found")
	}

	return tx.Commit(ctx)
}
