package group_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	group_models "github.com/svlynx/messenger/internal/chat/group/models"
)

type PostgresGroupRepo struct {
	db *pgxpool.Pool
}

func NewPostgresGroupRepo(db *pgxpool.Pool) *PostgresGroupRepo {
	return &PostgresGroupRepo{db: db}
}

func (r *PostgresGroupRepo) CreateGroupRepo(ctx context.Context, group *group_models.Group) (*group_models.Group, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO groups (id, name, handle, description, avatar_url, avatar_color, type, creator_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		group.ID, group.Name, group.Handle, group.Description,
		group.AvatarURL, group.AvatarColor, group.Type, group.CreatorID, group.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert group: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)`,
		group.ID, group.CreatorID, group_models.RoleCreator, group.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert creator member: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return group, nil
}

func (r *PostgresGroupRepo) GetGroupByIDRepo(ctx context.Context, groupID uuid.UUID) (*group_models.Group, error) {
	var g group_models.Group
	err := r.db.QueryRow(ctx, `
		SELECT id, name, handle, description, avatar_url, avatar_color, type, creator_id, created_at,
		       (SELECT COUNT(*) FROM group_members WHERE group_id = groups.id) AS member_count
		FROM groups WHERE id = $1`, groupID,
	).Scan(&g.ID, &g.Name, &g.Handle, &g.Description, &g.AvatarURL,
		&g.AvatarColor, &g.Type, &g.CreatorID, &g.CreatedAt, &g.MemberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get group by id: %w", err)
	}
	return &g, nil
}

func (r *PostgresGroupRepo) GetGroupByHandleRepo(ctx context.Context, handle string) (*group_models.Group, error) {
	var g group_models.Group
	err := r.db.QueryRow(ctx, `
		SELECT id, name, handle, description, avatar_url, avatar_color, type, creator_id, created_at,
		       (SELECT COUNT(*) FROM group_members WHERE group_id = groups.id) AS member_count
		FROM groups WHERE handle = $1`, handle,
	).Scan(&g.ID, &g.Name, &g.Handle, &g.Description, &g.AvatarURL,
		&g.AvatarColor, &g.Type, &g.CreatorID, &g.CreatedAt, &g.MemberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get group by handle: %w", err)
	}
	return &g, nil
}

func (r *PostgresGroupRepo) UpdateGroupRepo(ctx context.Context, group *group_models.Group) (*group_models.Group, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE groups
		SET name = $1, handle = $2, description = $3, avatar_url = $4, avatar_color = $5, type = $6
		WHERE id = $7`,
		group.Name, group.Handle, group.Description,
		group.AvatarURL, group.AvatarColor, group.Type, group.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}
	return group, nil
}

func (r *PostgresGroupRepo) DeleteGroupRepo(ctx context.Context, groupID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, q := range []string{
		`DELETE FROM group_invite_links WHERE group_id = $1`,
		`DELETE FROM group_messages WHERE group_id = $1`,
		`DELETE FROM group_topics WHERE group_id = $1`,
		`DELETE FROM group_bans WHERE group_id = $1`,
		`DELETE FROM group_members WHERE group_id = $1`,
		`DELETE FROM groups WHERE id = $1`,
	} {
		if _, err = tx.Exec(ctx, q, groupID); err != nil {
			return fmt.Errorf("delete group cascade: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresGroupRepo) SearchGroupsRepo(ctx context.Context, query string, limit int) ([]*group_models.Group, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, handle, description, avatar_url, avatar_color, type, creator_id, created_at,
		       (SELECT COUNT(*) FROM group_members WHERE group_id = groups.id) AS member_count
		FROM groups
		WHERE type = 'public'
		  AND (name ILIKE $1 OR handle ILIKE $1 OR description ILIKE $1)
		ORDER BY member_count DESC
		LIMIT $2`,
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search groups: %w", err)
	}
	defer rows.Close()
	return scanGroups(rows)
}

func (r *PostgresGroupRepo) AddMemberRepo(ctx context.Context, member *group_models.GroupMember) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, role, joined_at, invited_by, custom_name, is_banned)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (group_id, user_id) DO NOTHING`,
		member.GroupID, member.UserID, member.Role,
		member.JoinedAt, member.InvitedBy, member.CustomName, member.IsBanned,
	)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepo) RemoveMemberRepo(ctx context.Context, groupID, userID uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`,
		groupID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

func (r *PostgresGroupRepo) GetMemberRepo(ctx context.Context, groupID, userID uuid.UUID) (*group_models.GroupMember, error) {
	var m group_models.GroupMember
	err := r.db.QueryRow(ctx, `
		SELECT group_id, user_id, role, joined_at, invited_by, COALESCE(custom_name,''), is_banned, banned_until
		FROM group_members WHERE group_id = $1 AND user_id = $2`,
		groupID, userID,
	).Scan(&m.GroupID, &m.UserID, &m.Role, &m.JoinedAt, &m.InvitedBy, &m.CustomName, &m.IsBanned, &m.BannedUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get member: %w", err)
	}
	return &m, nil
}

func (r *PostgresGroupRepo) GetMembersRepo(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*group_models.GroupMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT group_id, user_id, role, joined_at, invited_by, COALESCE(custom_name,''), is_banned, banned_until
		FROM group_members WHERE group_id = $1
		ORDER BY
		  CASE role WHEN 'creator' THEN 1 WHEN 'admin' THEN 2 ELSE 3 END,
		  joined_at
		LIMIT $2 OFFSET $3`,
		groupID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	var members []*group_models.GroupMember
	for rows.Next() {
		var m group_models.GroupMember
		if err := rows.Scan(&m.GroupID, &m.UserID, &m.Role, &m.JoinedAt, &m.InvitedBy, &m.CustomName, &m.IsBanned, &m.BannedUntil); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, &m)
	}
	return members, rows.Err()
}

func (r *PostgresGroupRepo) UpdateMemberRoleRepo(ctx context.Context, groupID, userID uuid.UUID, role group_models.GroupRole) error {
	_, err := r.db.Exec(ctx, `
		UPDATE group_members SET role = $1 WHERE group_id = $2 AND user_id = $3`,
		role, groupID, userID,
	)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepo) GetUserGroupsRepo(ctx context.Context, userID uuid.UUID) ([]*group_models.GroupListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			g.id, g.name, g.handle, g.description,
			COALESCE(g.avatar_url, '') AS avatar_url,
			COALESCE(g.avatar_color, '') AS avatar_color,
			g.type,
			(SELECT COUNT(*) FROM group_members gm2 WHERE gm2.group_id = g.id) AS member_count,
			gm.role AS user_role,
			COALESCE(m.content, '') AS last_message_content,
			m.created_at AS last_message_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = $1
		LEFT JOIN LATERAL (
			SELECT content, created_at FROM group_messages
			WHERE group_id = g.id
			ORDER BY created_at DESC LIMIT 1
		) m ON true
		ORDER BY COALESCE(m.created_at, g.created_at) DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user groups: %w", err)
	}
	defer rows.Close()

	var list []*group_models.GroupListItem
	for rows.Next() {
		var item group_models.GroupListItem
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Handle, &item.Description,
			&item.AvatarURL, &item.AvatarColor, &item.Type,
			&item.MemberCount, &item.UserRole,
			&item.LastMessageContent, &item.LastMessageAt,
		); err != nil {
			return nil, fmt.Errorf("scan group list item: %w", err)
		}
		list = append(list, &item)
	}
	return list, rows.Err()
}

func (r *PostgresGroupRepo) GetMemberCountRepo(ctx context.Context, groupID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM group_members WHERE group_id = $1`, groupID,
	).Scan(&count)
	return count, err
}

func (r *PostgresGroupRepo) BanMemberRepo(ctx context.Context, ban *group_models.GroupBan) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO group_bans (group_id, user_id, banned_by, reason, banned_at, until)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (group_id, user_id) DO UPDATE SET banned_by = $3, reason = $4, banned_at = $5, until = $6`,
		ban.GroupID, ban.UserID, ban.BannedBy, ban.Reason, ban.BannedAt, ban.Until,
	)
	if err != nil {
		return fmt.Errorf("ban member: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		UPDATE group_members SET is_banned = true, banned_until = $3
		WHERE group_id = $1 AND user_id = $2`,
		ban.GroupID, ban.UserID, ban.Until,
	)
	if err != nil {
		return fmt.Errorf("mark member banned: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepo) UnbanMemberRepo(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM group_bans WHERE group_id = $1 AND user_id = $2`,
		groupID, userID,
	)
	if err != nil {
		return fmt.Errorf("unban member: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		UPDATE group_members SET is_banned = false, banned_until = NULL
		WHERE group_id = $1 AND user_id = $2`,
		groupID, userID,
	)
	if err != nil {
		return fmt.Errorf("mark member unbanned: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepo) IsBannedRepo(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var banned bool
	err := r.db.QueryRow(ctx, `
		SELECT is_banned FROM group_members WHERE group_id = $1 AND user_id = $2`,
		groupID, userID,
	).Scan(&banned)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check ban: %w", err)
	}

	if banned {
		var until *time.Time
		err := r.db.QueryRow(ctx, `
			SELECT banned_until FROM group_members WHERE group_id = $1 AND user_id = $2`,
			groupID, userID,
		).Scan(&until)
		if err != nil {
			return true, nil
		}
		if until != nil && until.Before(time.Now().UTC()) {
			return false, nil
		}
	}
	return banned, nil
}

func (r *PostgresGroupRepo) GetBansRepo(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*group_models.GroupBan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT group_id, user_id, banned_by, reason, banned_at, until
		FROM group_bans WHERE group_id = $1
		ORDER BY banned_at DESC
		LIMIT $2 OFFSET $3`,
		groupID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get bans: %w", err)
	}
	defer rows.Close()

	var bans []*group_models.GroupBan
	for rows.Next() {
		var b group_models.GroupBan
		if err := rows.Scan(&b.GroupID, &b.UserID, &b.BannedBy, &b.Reason, &b.BannedAt, &b.Until); err != nil {
			return nil, fmt.Errorf("scan ban: %w", err)
		}
		bans = append(bans, &b)
	}
	return bans, rows.Err()
}

func (r *PostgresGroupRepo) CreateTopicRepo(ctx context.Context, topic *group_models.GroupTopic) (*group_models.GroupTopic, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO group_topics (id, group_id, name, description, is_closed, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		topic.ID, topic.GroupID, topic.Name, topic.Description,
		topic.IsClosed, topic.CreatedBy, topic.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create topic: %w", err)
	}
	return topic, nil
}

func (r *PostgresGroupRepo) GetTopicByIDRepo(ctx context.Context, topicID uuid.UUID) (*group_models.GroupTopic, error) {
	var t group_models.GroupTopic
	err := r.db.QueryRow(ctx, `
		SELECT id, group_id, name, description, is_closed, created_by, created_at, updated_at
		FROM group_topics WHERE id = $1`, topicID,
	).Scan(&t.ID, &t.GroupID, &t.Name, &t.Description, &t.IsClosed, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get topic by id: %w", err)
	}
	return &t, nil
}

func (r *PostgresGroupRepo) GetTopicsByGroupRepo(ctx context.Context, groupID uuid.UUID) ([]*group_models.GroupTopic, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, group_id, name, description, is_closed, created_by, created_at, updated_at
		FROM group_topics WHERE group_id = $1
		ORDER BY created_at ASC`, groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("get topics by group: %w", err)
	}
	defer rows.Close()

	var topics []*group_models.GroupTopic
	for rows.Next() {
		var t group_models.GroupTopic
		if err := rows.Scan(&t.ID, &t.GroupID, &t.Name, &t.Description, &t.IsClosed, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan topic: %w", err)
		}
		topics = append(topics, &t)
	}
	return topics, rows.Err()
}

func (r *PostgresGroupRepo) UpdateTopicRepo(ctx context.Context, topic *group_models.GroupTopic) (*group_models.GroupTopic, error) {
	now := time.Now().UTC()
	topic.UpdatedAt = &now
	_, err := r.db.Exec(ctx, `
		UPDATE group_topics SET name = $1, description = $2, is_closed = $3, updated_at = $4
		WHERE id = $5`,
		topic.Name, topic.Description, topic.IsClosed, topic.UpdatedAt, topic.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update topic: %w", err)
	}
	return topic, nil
}

func (r *PostgresGroupRepo) DeleteTopicRepo(ctx context.Context, topicID uuid.UUID) error {
	res, err := r.db.Exec(ctx, `DELETE FROM group_topics WHERE id = $1`, topicID)
	if err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("topic not found")
	}
	return nil
}

func (r *PostgresGroupRepo) CreateGroupMessageRepo(ctx context.Context, msg *group_models.GroupMessage) (*group_models.GroupMessage, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO group_messages
			(id, group_id, topic_id, sender_id, content, type, status, media_url, media_type,
			 file_name, file_size, duration, transcript, reply_to_id, pinned, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		msg.ID, msg.GroupID, msg.TopicID, msg.SenderID, msg.Content,
		msg.Type, msg.Status, msg.MediaURL, msg.MediaType,
		msg.FileName, msg.FileSize, msg.Duration, msg.Transcript,
		msg.ReplyToID, msg.Pinned, msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create group message: %w", err)
	}
	return msg, nil
}

func (r *PostgresGroupRepo) GetGroupMessageByIDRepo(ctx context.Context, msgID uuid.UUID) (*group_models.GroupMessage, error) {
	var m group_models.GroupMessage
	err := r.db.QueryRow(ctx, `
		SELECT id, group_id, topic_id, sender_id, content, type, status,
		       COALESCE(media_url,''), COALESCE(media_type,''),
		       COALESCE(file_name,''), COALESCE(file_size,0),
		       COALESCE(duration,0), COALESCE(transcript,''),
		       reply_to_id, pinned, created_at, edited_at
		FROM group_messages WHERE id = $1`, msgID,
	).Scan(&m.ID, &m.GroupID, &m.TopicID, &m.SenderID, &m.Content, &m.Type, &m.Status,
		&m.MediaURL, &m.MediaType, &m.FileName, &m.FileSize,
		&m.Duration, &m.Transcript, &m.ReplyToID, &m.Pinned, &m.CreatedAt, &m.EditedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get group message by id: %w", err)
	}
	return &m, nil
}

func (r *PostgresGroupRepo) GetGroupMessagesRepo(ctx context.Context, groupID uuid.UUID, topicID *uuid.UUID, before time.Time, limit int) ([]*group_models.GroupMessage, error) {
	var rows pgx.Rows
	var err error
	if topicID != nil {
		rows, err = r.db.Query(ctx, `
			SELECT id, group_id, topic_id, sender_id, content, type, status,
			       COALESCE(media_url,''), COALESCE(media_type,''),
			       COALESCE(file_name,''), COALESCE(file_size,0),
			       COALESCE(duration,0), COALESCE(transcript,''),
			       reply_to_id, pinned, created_at, edited_at
			FROM group_messages
			WHERE group_id = $1 AND topic_id = $2 AND created_at < $3
			ORDER BY created_at DESC
			LIMIT $4`,
			groupID, *topicID, before, limit,
		)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, group_id, topic_id, sender_id, content, type, status,
			       COALESCE(media_url,''), COALESCE(media_type,''),
			       COALESCE(file_name,''), COALESCE(file_size,0),
			       COALESCE(duration,0), COALESCE(transcript,''),
			       reply_to_id, pinned, created_at, edited_at
			FROM group_messages
			WHERE group_id = $1 AND topic_id IS NULL AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3`,
			groupID, before, limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("get group messages: %w", err)
	}
	defer rows.Close()
	return scanGroupMessages(rows)
}

func (r *PostgresGroupRepo) UpdateGroupMessageRepo(ctx context.Context, msg *group_models.GroupMessage) (*group_models.GroupMessage, error) {
	now := time.Now().UTC()
	msg.EditedAt = &now
	_, err := r.db.Exec(ctx, `
		UPDATE group_messages
		SET content = $1, media_url = $2, media_type = $3, file_name = $4, file_size = $5, edited_at = $6
		WHERE id = $7`,
		msg.Content, msg.MediaURL, msg.MediaType, msg.FileName, msg.FileSize, msg.EditedAt, msg.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update group message: %w", err)
	}
	return msg, nil
}

func (r *PostgresGroupRepo) DeleteGroupMessageRepo(ctx context.Context, msgID uuid.UUID) error {
	res, err := r.db.Exec(ctx, `DELETE FROM group_messages WHERE id = $1`, msgID)
	if err != nil {
		return fmt.Errorf("delete group message: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("message not found")
	}
	return nil
}

func (r *PostgresGroupRepo) PinGroupMessageRepo(ctx context.Context, msgID uuid.UUID, pinned bool) error {
	_, err := r.db.Exec(ctx, `UPDATE group_messages SET pinned = $1 WHERE id = $2`, pinned, msgID)
	if err != nil {
		return fmt.Errorf("pin group message: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepo) GetPinnedGroupMessagesRepo(ctx context.Context, groupID uuid.UUID, topicID *uuid.UUID) ([]*group_models.GroupMessage, error) {
	var rows pgx.Rows
	var err error
	if topicID != nil {
		rows, err = r.db.Query(ctx, `
			SELECT id, group_id, topic_id, sender_id, content, type, status,
			       COALESCE(media_url,''), COALESCE(media_type,''),
			       COALESCE(file_name,''), COALESCE(file_size,0),
			       COALESCE(duration,0), COALESCE(transcript,''),
			       reply_to_id, pinned, created_at, edited_at
			FROM group_messages
			WHERE group_id = $1 AND topic_id = $2 AND pinned = true
			ORDER BY created_at DESC`,
			groupID, *topicID,
		)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, group_id, topic_id, sender_id, content, type, status,
			       COALESCE(media_url,''), COALESCE(media_type,''),
			       COALESCE(file_name,''), COALESCE(file_size,0),
			       COALESCE(duration,0), COALESCE(transcript,''),
			       reply_to_id, pinned, created_at, edited_at
			FROM group_messages
			WHERE group_id = $1 AND topic_id IS NULL AND pinned = true
			ORDER BY created_at DESC`,
			groupID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("get pinned group messages: %w", err)
	}
	defer rows.Close()
	return scanGroupMessages(rows)
}

func (r *PostgresGroupRepo) SearchGroupMessagesRepo(ctx context.Context, groupID uuid.UUID, query string) ([]*group_models.GroupMessage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, group_id, topic_id, sender_id, content, type, status,
		       COALESCE(media_url,''), COALESCE(media_type,''),
		       COALESCE(file_name,''), COALESCE(file_size,0),
		       COALESCE(duration,0), COALESCE(transcript,''),
		       reply_to_id, pinned, created_at, edited_at
		FROM group_messages
		WHERE group_id = $1 AND content ILIKE $2
		ORDER BY created_at DESC`,
		groupID, "%"+query+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("search group messages: %w", err)
	}
	defer rows.Close()
	return scanGroupMessages(rows)
}

func (r *PostgresGroupRepo) CreateInviteLinkRepo(ctx context.Context, link *group_models.GroupInviteLink) (*group_models.GroupInviteLink, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO group_invite_links
			(id, group_id, created_by, token, expires_at, max_uses, use_count, created_at, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		link.ID, link.GroupID, link.CreatedBy, link.Token,
		link.ExpiresAt, link.MaxUses, link.UseCount, link.CreatedAt, link.Active,
	)
	if err != nil {
		return nil, fmt.Errorf("create invite link: %w", err)
	}
	return link, nil
}

func (r *PostgresGroupRepo) GetInviteLinkByTokenRepo(ctx context.Context, token string) (*group_models.GroupInviteLink, error) {
	var l group_models.GroupInviteLink
	err := r.db.QueryRow(ctx, `
		SELECT id, group_id, created_by, token, expires_at, max_uses, use_count, created_at, active
		FROM group_invite_links WHERE token = $1`, token,
	).Scan(&l.ID, &l.GroupID, &l.CreatedBy, &l.Token,
		&l.ExpiresAt, &l.MaxUses, &l.UseCount, &l.CreatedAt, &l.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get invite link: %w", err)
	}
	return &l, nil
}

func (r *PostgresGroupRepo) GetInviteLinksByGroupRepo(ctx context.Context, groupID uuid.UUID) ([]*group_models.GroupInviteLink, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, group_id, created_by, token, expires_at, max_uses, use_count, created_at, active
		FROM group_invite_links WHERE group_id = $1
		ORDER BY created_at DESC`, groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("get invite links: %w", err)
	}
	defer rows.Close()

	var links []*group_models.GroupInviteLink
	for rows.Next() {
		var l group_models.GroupInviteLink
		if err := rows.Scan(&l.ID, &l.GroupID, &l.CreatedBy, &l.Token,
			&l.ExpiresAt, &l.MaxUses, &l.UseCount, &l.CreatedAt, &l.Active); err != nil {
			return nil, fmt.Errorf("scan invite link: %w", err)
		}
		links = append(links, &l)
	}
	return links, rows.Err()
}

func (r *PostgresGroupRepo) IncrementInviteUsageRepo(ctx context.Context, linkID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE group_invite_links SET use_count = use_count + 1 WHERE id = $1`, linkID)
	return err
}

func (r *PostgresGroupRepo) DeactivateInviteLinkRepo(ctx context.Context, linkID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE group_invite_links SET active = false WHERE id = $1`, linkID)
	return err
}

func scanGroups(rows pgx.Rows) ([]*group_models.Group, error) {
	var result []*group_models.Group
	for rows.Next() {
		var g group_models.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Handle, &g.Description,
			&g.AvatarURL, &g.AvatarColor, &g.Type, &g.CreatorID, &g.CreatedAt, &g.MemberCount); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		result = append(result, &g)
	}
	return result, rows.Err()
}

func scanGroupMessages(rows pgx.Rows) ([]*group_models.GroupMessage, error) {
	var msgs []*group_models.GroupMessage
	for rows.Next() {
		var m group_models.GroupMessage
		if err := rows.Scan(&m.ID, &m.GroupID, &m.TopicID, &m.SenderID, &m.Content,
			&m.Type, &m.Status, &m.MediaURL, &m.MediaType,
			&m.FileName, &m.FileSize, &m.Duration, &m.Transcript,
			&m.ReplyToID, &m.Pinned, &m.CreatedAt, &m.EditedAt); err != nil {
			return nil, fmt.Errorf("scan group message: %w", err)
		}
		msgs = append(msgs, &m)
	}
	return msgs, rows.Err()
}