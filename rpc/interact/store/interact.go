// package store 封装互动相关数据表（likes/favorites/comments/danmakus）的 CRUD 操作。
// 使用 go-zero sqlx 作为轻量 ORM。
package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// InteractStore 封装四张互动表的所有数据库操作。
type InteractStore struct {
	conn sqlx.SqlConn
}

func NewInteractStore(conn sqlx.SqlConn) *InteractStore {
	return &InteractStore{conn: conn}
}

// ─────────── 点赞 ───────────

func (s *InteractStore) InsertLike(ctx context.Context, userId, videoId int64) error {
	_, err := s.conn.ExecCtx(ctx, `
		INSERT INTO likes (user_id, video_id, created_at) VALUES (?, ?, NOW())
		ON DUPLICATE KEY UPDATE created_at = NOW()
	`, userId, videoId)
	if err != nil {
		return err
	}
	_, _ = s.conn.ExecCtx(ctx, `UPDATE videos SET like_count = like_count + 1 WHERE id = ?`, videoId)
	return nil
}

func (s *InteractStore) DeleteLike(ctx context.Context, userId, videoId int64) error {
	result, err := s.conn.ExecCtx(ctx, `
		DELETE FROM likes WHERE user_id = ? AND video_id = ?
	`, userId, videoId)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		_, _ = s.conn.ExecCtx(ctx, `UPDATE videos SET like_count = GREATEST(like_count - 1, 0) WHERE id = ?`, videoId)
	}
	return nil
}

// IsLiked 查询用户是否已点赞某个视频。
func (s *InteractStore) IsLiked(ctx context.Context, userId, videoId int64) (bool, error) {
	var count int
	err := s.conn.QueryRowCtx(ctx, &count, `
		SELECT COUNT(1) FROM likes WHERE user_id = ? AND video_id = ?
	`, userId, videoId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountLikes 查询视频的点赞总数。
func (s *InteractStore) CountLikes(ctx context.Context, videoId int64) int64 {
	var count int64
	_ = s.conn.QueryRowCtx(ctx, &count, `SELECT COUNT(1) FROM likes WHERE video_id = ?`, videoId)
	return count
}

// ─────────── 收藏 ───────────

func (s *InteractStore) InsertFavorite(ctx context.Context, userId, videoId int64) error {
	_, err := s.conn.ExecCtx(ctx, `
		INSERT INTO favorites (user_id, video_id, created_at) VALUES (?, ?, NOW())
		ON DUPLICATE KEY UPDATE created_at = NOW()
	`, userId, videoId)
	return err
}

func (s *InteractStore) DeleteFavorite(ctx context.Context, userId, videoId int64) error {
	_, err := s.conn.ExecCtx(ctx, `
		DELETE FROM favorites WHERE user_id = ? AND video_id = ?
	`, userId, videoId)
	return err
}

func (s *InteractStore) IsFavorited(ctx context.Context, userId, videoId int64) (bool, error) {
	var count int
	err := s.conn.QueryRowCtx(ctx, &count, `
		SELECT COUNT(1) FROM favorites WHERE user_id = ? AND video_id = ?
	`, userId, videoId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ─────────── 评论 ───────────

// CommentRow 查询返回的评论行（不 join users，username/avatar 由调用方填充）。
type CommentRow struct {
	Id         int64  `db:"id"`
	UserId     int64  `db:"user_id"`
	VideoId    int64  `db:"video_id"`
	ParentId   int64  `db:"parent_id"`
	Content    string `db:"content"`
	LikeCount  int64  `db:"like_count"`
	ReplyCount int    `db:"reply_count"`
	CreatedAt  int64  `db:"created_at"` // Unix 时间戳
}

func (s *InteractStore) InsertComment(ctx context.Context, userId, videoId, parentId int64, content string) (int64, error) {
	result, err := s.conn.ExecCtx(ctx, `
		INSERT INTO comments (user_id, video_id, parent_id, content, created_at)
		VALUES (?, ?, ?, ?, NOW())
	`, userId, videoId, parentId, content)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ListComments 游标分页查询评论列表。
func (s *InteractStore) ListComments(ctx context.Context, videoId, cursor int64, limit int32, sort string) ([]*CommentRow, error) {
	var rows []*CommentRow

	query := " WHERE video_id = ? AND parent_id = 0" // 先只查顶级评论
	args := []any{videoId}

	if cursor > 0 {
		if sort == "hot" {
			query += " AND like_count < ?"
		} else {
			query += " AND id < ?"
		}
		args = append(args, cursor)
	}

	order := " ORDER BY id DESC"
	if sort == "hot" {
		order = " ORDER BY like_count DESC, id DESC"
	}

	query += order + " LIMIT ?"
	args = append(args, limit+1)

	err := s.conn.QueryRowsCtx(ctx, &rows, `
		SELECT id, user_id, video_id, parent_id, content, like_count, reply_count,
			   UNIX_TIMESTAMP(created_at) AS created_at
		FROM comments `+query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *InteractStore) DeleteComment(ctx context.Context, commentId, userId int64) error {
	_, err := s.conn.ExecCtx(ctx, `
		DELETE FROM comments WHERE id = ? AND user_id = ?
	`, commentId, userId)
	return err
}

// ─────────── 弹幕 ───────────

type DanmakuRow struct {
	Id        int64   `db:"id"`
	UserId    int64   `db:"user_id"`
	VideoId   int64   `db:"video_id"`
	Content   string  `db:"content"`
	Time      float64 `db:"time"`
	Color     string  `db:"color"`
	Mode      int32   `db:"mode"`
	CreatedAt int64   `db:"created_at"`
}

func (s *InteractStore) InsertDanmaku(ctx context.Context, userId, videoId int64, content string, time float64, color string, mode int32) (int64, error) {
	result, err := s.conn.ExecCtx(ctx, `
		INSERT INTO danmakus (user_id, video_id, content, time, color, mode, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`, userId, videoId, content, time, color, mode)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetDanmakus 获取某个视频在指定时间点附近的弹幕（前后各 5 秒）。
func (s *InteractStore) GetDanmakus(ctx context.Context, videoId int64, time float64) ([]*DanmakuRow, error) {
	var rows []*DanmakuRow
	err := s.conn.QueryRowsCtx(ctx, &rows, `
		SELECT id, user_id, video_id, content, time, color, mode,
			   UNIX_TIMESTAMP(created_at) AS created_at
		FROM danmakus
		WHERE video_id = ? AND time BETWEEN ? AND ?
		ORDER BY time ASC
		LIMIT 200
	`, videoId, time-5, time+5)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
