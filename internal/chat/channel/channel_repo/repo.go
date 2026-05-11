package channelrepo


import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	channel_models "github.com/svlynx/messenger/internal/chat/channel/channel_models"
)

type PostgresChannelRepo struct {
	db *pgxpool.Pool
}

func NewPostgresChannelRepo(db *pgxpool.Pool) *PostgresChannelRepo {
	return &PostgresChannelRepo{db: db}
}


func (r *PostgresChannelRepo) CreateChannelRepo(ctx context.Context, channel *channel_models.Channel) (*channel_models.Channel, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO channels (id, name, handle, description, avatar_url, avatar_color, type, owner_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		channel.ID, channel.Name, channel.Handle, channel.Description,
		channel.AvatarURL, channel.AvatarColor, channel.Type, channel.OwnerID, channel.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert channel: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO channel_members (channel_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)`,
		channel.ID, channel.OwnerID, channel_models.RoleOwner, channel.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert owner member: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return channel, nil
}

func (r *PostgresChannelRepo) GetChannelByIDRepo(ctx context.Context, channelID uuid.UUID) (*channel_models.Channel, error) {
	var ch channel_models.Channel
	err := r.db.QueryRow(ctx, `
		SELECT id, name, handle, description, avatar_url, avatar_color, type, owner_id, created_at,
		       (SELECT COUNT(*) FROM channel_members WHERE channel_id = channels.id) AS member_count
		FROM channels WHERE id = $1`, channelID,
	).Scan(&ch.ID, &ch.Name, &ch.Handle, &ch.Description, &ch.AvatarURL,
		&ch.AvatarColor, &ch.Type, &ch.OwnerID, &ch.CreatedAt, &ch.MemberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get channel by id: %w", err)
	}
	return &ch, nil
}

func (r *PostgresChannelRepo) GetChannelByHandleRepo(ctx context.Context, handle string) (*channel_models.Channel, error) {
	var ch channel_models.Channel
	err := r.db.QueryRow(ctx, `
		SELECT id, name, handle, description, avatar_url, avatar_color, type, owner_id, created_at,
		       (SELECT COUNT(*) FROM channel_members WHERE channel_id = channels.id) AS member_count
		FROM channels WHERE handle = $1`, handle,
	).Scan(&ch.ID, &ch.Name, &ch.Handle, &ch.Description, &ch.AvatarURL,
		&ch.AvatarColor, &ch.Type, &ch.OwnerID, &ch.CreatedAt, &ch.MemberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get channel by handle: %w", err)
	}
	return &ch, nil
}

func (r *PostgresChannelRepo) UpdateChannelRepo(ctx context.Context, channel *channel_models.Channel) (*channel_models.Channel, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE channels
		SET name = $1, handle = $2, description = $3, avatar_url = $4, avatar_color = $5, type = $6
		WHERE id = $7`,
		channel.Name, channel.Handle, channel.Description,
		channel.AvatarURL, channel.AvatarColor, channel.Type, channel.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update channel: %w", err)
	}
	return channel, nil
}

func (r *PostgresChannelRepo) DeleteChannelRepo(ctx context.Context, channelID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, q := range []string{
		`DELETE FROM channel_invite_links WHERE channel_id = $1`,
		`DELETE FROM channel_posts WHERE channel_id = $1`,
		`DELETE FROM channel_members WHERE channel_id = $1`,
		`DELETE FROM channels WHERE id = $1`,
	} {
		if _, err = tx.Exec(ctx, q, channelID); err != nil {
			return fmt.Errorf("delete channel cascade: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresChannelRepo) SearchChannelsRepo(ctx context.Context, query string, limit int) ([]*channel_models.Channel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, handle, description, avatar_url, avatar_color, type, owner_id, created_at,
		       (SELECT COUNT(*) FROM channel_members WHERE channel_id = channels.id) AS member_count
		FROM channels
		WHERE type = 'public'
		  AND (name ILIKE $1 OR handle ILIKE $1 OR description ILIKE $1)
		ORDER BY member_count DESC
		LIMIT $2`,
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search channels: %w", err)
	}
	defer rows.Close()
	return scanChannels(rows)
}


func (r *PostgresChannelRepo) AddMemberRepo(ctx context.Context, member *channel_models.ChannelMember) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO channel_members (channel_id, user_id, role, joined_at, invited_by, custom_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (channel_id, user_id) DO NOTHING`,
		member.ChannelID, member.UserID, member.Role,
		member.JoinedAt, member.InvitedBy, member.CustomName,
	)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *PostgresChannelRepo) RemoveMemberRepo(ctx context.Context, channelID, userID uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

func (r *PostgresChannelRepo) GetMemberRepo(ctx context.Context, channelID, userID uuid.UUID) (*channel_models.ChannelMember, error) {
	var m channel_models.ChannelMember
	err := r.db.QueryRow(ctx, `
		SELECT channel_id, user_id, role, joined_at, invited_by, COALESCE(custom_name,'')
		FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelID, userID,
	).Scan(&m.ChannelID, &m.UserID, &m.Role, &m.JoinedAt, &m.InvitedBy, &m.CustomName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get member: %w", err)
	}
	return &m, nil
}

func (r *PostgresChannelRepo) GetMembersRepo(ctx context.Context, channelID uuid.UUID, limit, offset int) ([]*channel_models.ChannelMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT channel_id, user_id, role, joined_at, invited_by, COALESCE(custom_name,'')
		FROM channel_members WHERE channel_id = $1
		ORDER BY
		  CASE role WHEN 'owner' THEN 1 WHEN 'admin' THEN 2 WHEN 'editor' THEN 3 ELSE 4 END,
		  joined_at
		LIMIT $2 OFFSET $3`,
		channelID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	var members []*channel_models.ChannelMember
	for rows.Next() {
		var m channel_models.ChannelMember
		if err := rows.Scan(&m.ChannelID, &m.UserID, &m.Role, &m.JoinedAt, &m.InvitedBy, &m.CustomName); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, &m)
	}
	return members, rows.Err()
}

func (r *PostgresChannelRepo) UpdateMemberRoleRepo(ctx context.Context, channelID, userID uuid.UUID, role channel_models.ChannelRole) error {
	_, err := r.db.Exec(ctx, `
		UPDATE channel_members SET role = $1 WHERE channel_id = $2 AND user_id = $3`,
		role, channelID, userID,
	)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	return nil
}

func (r *PostgresChannelRepo) GetUserChannelsRepo(ctx context.Context, userID uuid.UUID) ([]*channel_models.ChannelListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			c.id, c.name, c.handle, c.description,
			COALESCE(c.avatar_url, '') AS avatar_url,
			COALESCE(c.avatar_color, '') AS avatar_color,
			c.type,
			(SELECT COUNT(*) FROM channel_members cm2 WHERE cm2.channel_id = c.id) AS member_count,
			cm.role AS user_role,
			COALESCE(p.content, '') AS last_post_content,
			p.created_at AS last_post_at
		FROM channels c
		JOIN channel_members cm ON cm.channel_id = c.id AND cm.user_id = $1
		LEFT JOIN LATERAL (
			SELECT content, created_at FROM channel_posts
			WHERE channel_id = c.id
			ORDER BY created_at DESC LIMIT 1
		) p ON true
		ORDER BY COALESCE(p.created_at, c.created_at) DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user channels: %w", err)
	}
	defer rows.Close()

	var list []*channel_models.ChannelListItem
	for rows.Next() {
		var item channel_models.ChannelListItem
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Handle, &item.Description,
			&item.AvatarURL, &item.AvatarColor, &item.Type,
			&item.MemberCount, &item.UserRole,
			&item.LastPostContent, &item.LastPostAt,
		); err != nil {
			return nil, fmt.Errorf("scan channel list item: %w", err)
		}
		list = append(list, &item)
	}
	return list, rows.Err()
}

func (r *PostgresChannelRepo) GetMemberCountRepo(ctx context.Context, channelID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM channel_members WHERE channel_id = $1`, channelID,
	).Scan(&count)
	return count, err
}


func (r *PostgresChannelRepo) CreatePostRepo(ctx context.Context, post *channel_models.ChannelPost) (*channel_models.ChannelPost, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO channel_posts
			(id, channel_id, author_id, content, media_url, media_type, file_name, file_size, pinned, view_count, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		post.ID, post.ChannelID, post.AuthorID, post.Content,
		post.MediaURL, post.MediaType, post.FileName, post.FileSize,
		post.Pinned, post.ViewCount, post.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	return post, nil
}

func (r *PostgresChannelRepo) GetPostByIDRepo(ctx context.Context, postID uuid.UUID) (*channel_models.ChannelPost, error) {
	var p channel_models.ChannelPost
	err := r.db.QueryRow(ctx, `
		SELECT id, channel_id, author_id, content, COALESCE(media_url,''), COALESCE(media_type,''),
		       COALESCE(file_name,''), COALESCE(file_size,0), pinned, view_count, created_at, edited_at
		FROM channel_posts WHERE id = $1`, postID,
	).Scan(&p.ID, &p.ChannelID, &p.AuthorID, &p.Content, &p.MediaURL, &p.MediaType,
		&p.FileName, &p.FileSize, &p.Pinned, &p.ViewCount, &p.CreatedAt, &p.EditedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get post by id: %w", err)
	}
	return &p, nil
}

func (r *PostgresChannelRepo) GetPostsByChannelRepo(ctx context.Context, channelID uuid.UUID, before time.Time, limit int) ([]*channel_models.ChannelPost, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, channel_id, author_id, content, COALESCE(media_url,''), COALESCE(media_type,''),
		       COALESCE(file_name,''), COALESCE(file_size,0), pinned, view_count, created_at, edited_at
		FROM channel_posts
		WHERE channel_id = $1 AND created_at < $2
		ORDER BY created_at DESC
		LIMIT $3`,
		channelID, before, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get posts: %w", err)
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *PostgresChannelRepo) UpdatePostRepo(ctx context.Context, post *channel_models.ChannelPost) (*channel_models.ChannelPost, error) {
	now := time.Now()
	post.EditedAt = &now
	_, err := r.db.Exec(ctx, `
		UPDATE channel_posts
		SET content = $1, media_url = $2, media_type = $3, file_name = $4, file_size = $5, edited_at = $6
		WHERE id = $7`,
		post.Content, post.MediaURL, post.MediaType, post.FileName, post.FileSize, post.EditedAt, post.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}
	return post, nil
}

func (r *PostgresChannelRepo) DeletePostRepo(ctx context.Context, postID uuid.UUID) error {
	res, err := r.db.Exec(ctx, `DELETE FROM channel_posts WHERE id = $1`, postID)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

func (r *PostgresChannelRepo) PinPostRepo(ctx context.Context, postID uuid.UUID, pinned bool) error {
	_, err := r.db.Exec(ctx, `UPDATE channel_posts SET pinned = $1 WHERE id = $2`, pinned, postID)
	if err != nil {
		return fmt.Errorf("pin post: %w", err)
	}
	return nil
}

func (r *PostgresChannelRepo) GetPinnedPostsRepo(ctx context.Context, channelID uuid.UUID) ([]*channel_models.ChannelPost, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, channel_id, author_id, content, COALESCE(media_url,''), COALESCE(media_type,''),
		       COALESCE(file_name,''), COALESCE(file_size,0), pinned, view_count, created_at, edited_at
		FROM channel_posts
		WHERE channel_id = $1 AND pinned = true
		ORDER BY created_at DESC`,
		channelID,
	)
	if err != nil {
		return nil, fmt.Errorf("get pinned posts: %w", err)
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *PostgresChannelRepo) SearchPostsRepo(ctx context.Context, channelID uuid.UUID, query string) ([]*channel_models.ChannelPost, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, channel_id, author_id, content, COALESCE(media_url,''), COALESCE(media_type,''),
		       COALESCE(file_name,''), COALESCE(file_size,0), pinned, view_count, created_at, edited_at
		FROM channel_posts
		WHERE channel_id = $1 AND content ILIKE $2
		ORDER BY created_at DESC`,
		channelID, "%"+query+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *PostgresChannelRepo) IncrementViewCountRepo(ctx context.Context, postID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE channel_posts SET view_count = view_count + 1 WHERE id = $1`, postID)
	return err
}


func (r *PostgresChannelRepo) CreateInviteLinkRepo(ctx context.Context, link *channel_models.ChannelInviteLink) (*channel_models.ChannelInviteLink, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO channel_invite_links
			(id, channel_id, created_by, token, expires_at, max_uses, use_count, created_at, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		link.ID, link.ChannelID, link.CreatedBy, link.Token,
		link.ExpiresAt, link.MaxUses, link.UseCount, link.CreatedAt, link.Active,
	)
	if err != nil {
		return nil, fmt.Errorf("create invite link: %w", err)
	}
	return link, nil
}

func (r *PostgresChannelRepo) GetInviteLinkByTokenRepo(ctx context.Context, token string) (*channel_models.ChannelInviteLink, error) {
	var l channel_models.ChannelInviteLink
	err := r.db.QueryRow(ctx, `
		SELECT id, channel_id, created_by, token, expires_at, max_uses, use_count, created_at, active
		FROM channel_invite_links WHERE token = $1`, token,
	).Scan(&l.ID, &l.ChannelID, &l.CreatedBy, &l.Token,
		&l.ExpiresAt, &l.MaxUses, &l.UseCount, &l.CreatedAt, &l.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get invite link: %w", err)
	}
	return &l, nil
}

func (r *PostgresChannelRepo) GetInviteLinksByChannelRepo(ctx context.Context, channelID uuid.UUID) ([]*channel_models.ChannelInviteLink, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, channel_id, created_by, token, expires_at, max_uses, use_count, created_at, active
		FROM channel_invite_links WHERE channel_id = $1
		ORDER BY created_at DESC`, channelID,
	)
	if err != nil {
		return nil, fmt.Errorf("get invite links: %w", err)
	}
	defer rows.Close()

	var links []*channel_models.ChannelInviteLink
	for rows.Next() {
		var l channel_models.ChannelInviteLink
		if err := rows.Scan(&l.ID, &l.ChannelID, &l.CreatedBy, &l.Token,
			&l.ExpiresAt, &l.MaxUses, &l.UseCount, &l.CreatedAt, &l.Active); err != nil {
			return nil, fmt.Errorf("scan invite link: %w", err)
		}
		links = append(links, &l)
	}
	return links, rows.Err()
}

func (r *PostgresChannelRepo) IncrementInviteUsageRepo(ctx context.Context, linkID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE channel_invite_links SET use_count = use_count + 1 WHERE id = $1`, linkID)
	return err
}

func (r *PostgresChannelRepo) DeactivateInviteLinkRepo(ctx context.Context, linkID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE channel_invite_links SET active = false WHERE id = $1`, linkID)
	return err
}

func scanChannels(rows pgx.Rows) ([]*channel_models.Channel, error) {
	var result []*channel_models.Channel
	for rows.Next() {
		var ch channel_models.Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Handle, &ch.Description,
			&ch.AvatarURL, &ch.AvatarColor, &ch.Type, &ch.OwnerID, &ch.CreatedAt, &ch.MemberCount); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		result = append(result, &ch)
	}
	return result, rows.Err()
}

func scanPosts(rows pgx.Rows) ([]*channel_models.ChannelPost, error) {
	var posts []*channel_models.ChannelPost
	for rows.Next() {
		var p channel_models.ChannelPost
		if err := rows.Scan(&p.ID, &p.ChannelID, &p.AuthorID, &p.Content,
			&p.MediaURL, &p.MediaType, &p.FileName, &p.FileSize,
			&p.Pinned, &p.ViewCount, &p.CreatedAt, &p.EditedAt); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, &p)
	}
	return posts, rows.Err()
}