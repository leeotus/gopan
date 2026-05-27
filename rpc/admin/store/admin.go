package store

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminStore struct {
	conn sqlx.SqlConn
}

func NewAdminStore(conn sqlx.SqlConn) *AdminStore {
	return &AdminStore{conn: conn}
}

type UserRow struct {
	Id       int64  `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Role     int    `db:"role"`
}

func (s *AdminStore) FindAdminByUsername(ctx context.Context, username string) (*UserRow, error) {
	var u UserRow
	err := s.conn.QueryRowCtx(ctx, &u, `
		SELECT id, username, password, role FROM users WHERE username = ? AND role = 1 AND deleted_at IS NULL
	`, username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type VideoRow struct {
	Id        int64  `db:"id"`
	Title     string `db:"title"`
	CoverUrl  string `db:"cover_url"`
	UserId    int64  `db:"user_id"`
	Status    int32  `db:"status"`
	PlayCount int64  `db:"play_count"`
	CreatedAt int64  `db:"created_at"`
}

func (s *AdminStore) ListVideos(ctx context.Context, cursor int64, limit int32, status int32) ([]*VideoRow, error) {
	query := "WHERE deleted_at IS NULL"
	args := []any{}
	if status >= 0 {
		query += " AND status = ?"
		args = append(args, status)
	}
	if cursor > 0 {
		query += " AND id < ?"
		args = append(args, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit+1)

	var rows []*VideoRow
	err := s.conn.QueryRowsCtx(ctx, &rows, `
		SELECT id, title, cover_url, user_id, status, play_count, UNIX_TIMESTAMP(created_at) AS created_at
		FROM videos `+query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return rows, nil
}

func (s *AdminStore) UpdateVideoStatus(ctx context.Context, videoId int64, status int32) error {
	_, err := s.conn.ExecCtx(ctx, `UPDATE videos SET status = ? WHERE id = ?`, status, videoId)
	return err
}
