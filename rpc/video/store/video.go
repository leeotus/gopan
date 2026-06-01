// package store 封装 videos 表和 transcodes 表的数据库操作。
// 采用游标分页（cursor-based pagination）代替传统 offset 分页，
// 在大数据量下性能更稳定。
package store

import (
	"context"
	"database/sql"
	"fmt"

	"gopan/rpc/video/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// VideoStore 封装视频相关的所有 DB 操作。
type VideoStore struct {
	conn sqlx.SqlConn
}

// NewVideoStore 创建 VideoStore。
func NewVideoStore(conn sqlx.SqlConn) *VideoStore {
	return &VideoStore{conn: conn}
}

// Insert 插入一条视频记录（包含断点上传的 total_chunks 和 upload_id）。
func (s *VideoStore) Insert(ctx context.Context, v *model.Video) (sql.Result, error) {
	return s.conn.ExecCtx(ctx, `
		INSERT INTO videos (title, description, user_id, object_key, cover_url, category,
			duration, file_size, file_hash, total_chunks, upload_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, v.Title, v.Description, v.UserId, v.ObjectKey, v.CoverUrl, v.Category,
		v.Duration, v.FileSize, v.FileHash, v.TotalChunks, v.UploadId, v.Status)
}

// FindById 根据主键查找视频（排除软删除记录）。
func (s *VideoStore) FindById(ctx context.Context, id int64) (*model.Video, error) {
	var v model.Video
	err := s.conn.QueryRowCtx(ctx, &v, `
		SELECT id, title, description, user_id, object_key, cover_url, category,
			duration, file_size, file_hash, total_chunks, upload_id, status, play_count, like_count, created_at, updated_at, deleted_at
		FROM videos WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// List 视频列表，支持游标分页、分类过滤和排序。
// cursor=0 表示第一页；category="" 表示全部；
// sort="hot" 按播放量排序，其他按创建时间倒序。
// 多查一条数据以判断 has_more（是否存在下一页）。
func (s *VideoStore) List(ctx context.Context, cursor int64, limit int32, category, sort string) ([]*model.Video, error) {
	var (
		videos []*model.Video
		query  string
		args   []any
	)

	if cursor > 0 {
		query = "WHERE id < ? AND deleted_at IS NULL"
		args = append(args, cursor)
	} else {
		query = "WHERE deleted_at IS NULL"
	}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}

	query += " AND status = 2" // 只展示转码完成的视频

	orderBy := "ORDER BY id DESC"
	if sort == "hot" {
		orderBy = "ORDER BY play_count DESC, id DESC"
	}

	query += " " + orderBy + " LIMIT ?"
	args = append(args, limit+1) // 多查一条用于 has_more 判断

	err := s.conn.QueryRowsCtx(ctx, &videos, fmt.Sprintf(`
		SELECT id, title, description, user_id, object_key, cover_url, category,
			duration, file_size, file_hash, total_chunks, upload_id, status, play_count, like_count, created_at, updated_at, deleted_at
		FROM videos %s
	`, query), args...)
	if err != nil {
		return nil, err
	}
	return videos, nil
}

// ListByUser 获取指定用户上传的视频列表。
func (s *VideoStore) ListByUser(ctx context.Context, userId, cursor int64, limit int32) ([]*model.Video, error) {
	var videos []*model.Video
	query := "WHERE user_id = ? AND deleted_at IS NULL"
	args := []any{userId}

	if cursor > 0 {
		query += " AND id < ?"
		args = append(args, cursor)
	}

	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit+1)

	err := s.conn.QueryRowsCtx(ctx, &videos, fmt.Sprintf(`
		SELECT id, title, description, user_id, object_key, cover_url, category,
			duration, file_size, file_hash, total_chunks, upload_id, status, play_count, like_count, created_at, updated_at, deleted_at
		FROM videos %s
	`, query), args...)
	if err != nil {
		return nil, err
	}
	return videos, nil
}

// Update 更新视频的标题、简介和分类。
func (s *VideoStore) Update(ctx context.Context, v *model.Video) error {
	_, err := s.conn.ExecCtx(ctx, `
		UPDATE videos SET title = ?, description = ?, category = ?, updated_at = NOW()
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`, v.Title, v.Description, v.Category, v.Id, v.UserId)
	return err
}

// Delete 软删除视频，设置 deleted_at = NOW()，不会物理删除数据。
// 仅允许视频所有者删除自己的视频。
func (s *VideoStore) Delete(ctx context.Context, id, userId int64) error {
	_, err := s.conn.ExecCtx(ctx, `
		UPDATE videos SET deleted_at = NOW() WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`, id, userId)
	return err
}

// UpdateTranscode 转码完成后回调，更新视频状态、封面和时长。
func (s *VideoStore) UpdateTranscode(ctx context.Context, videoId int64, status int32, coverUrl string, duration int32) error {
	_, err := s.conn.ExecCtx(ctx, `
		UPDATE videos SET status = ?, cover_url = ?, duration = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, status, coverUrl, duration, videoId)
	return err
}

// UpdateStatus 更新视频状态（如标记为转码中、正常、失败等）。
func (s *VideoStore) UpdateStatus(ctx context.Context, videoId int64, status int32) error {
	_, err := s.conn.ExecCtx(ctx, `
		UPDATE videos SET status = ?, updated_at = NOW() WHERE id = ?
	`, status, videoId)
	return err
}

// InsertTranscode 插入一条转码结果记录（某个分辨率的 HLS 流）。
func (s *VideoStore) InsertTranscode(ctx context.Context, t *model.Transcode) error {
	_, err := s.conn.ExecCtx(ctx, `
		INSERT INTO transcodes (video_id, resolution, m3u8_url, bitrate)
		VALUES (?, ?, ?, ?)
	`, t.VideoId, t.Resolution, t.M3U8Url, t.Bitrate)
	return err
}

// GetTranscodes 获取某个视频的所有转码信息（多分辨率 HLS 地址）。
func (s *VideoStore) GetTranscodes(ctx context.Context, videoId int64) ([]*model.Transcode, error) {
	var transcodes []*model.Transcode
	err := s.conn.QueryRowsCtx(ctx, &transcodes, `
		SELECT id, video_id, resolution, m3u8_url, bitrate
		FROM transcodes WHERE video_id = ?
	`, videoId)
	if err != nil {
		return nil, err
	}
	return transcodes, nil
}

// IncrPlayCount 原子增加播放计数（直接 UPDATE SET play_count = play_count + 1）。
func (s *VideoStore) IncrPlayCount(ctx context.Context, videoId int64) error {
	_, err := s.conn.ExecCtx(ctx, `
		UPDATE videos SET play_count = play_count + 1 WHERE id = ?
	`, videoId)
	return err
}

// IncrLikeCount 原子增减点赞计数，delta 可以是 +1 或 -1。
func (s *VideoStore) IncrLikeCount(ctx context.Context, videoId int64, delta int64) error {
	_, err := s.conn.ExecCtx(ctx, `
		UPDATE videos SET like_count = like_count + ? WHERE id = ?
	`, delta, videoId)
	return err
}

// FindByUploadId 根据 upload_id 查找视频（断点上传状态恢复用）。
func (s *VideoStore) FindByUploadId(ctx context.Context, uploadId string) (*model.Video, error) {
	var v model.Video
	err := s.conn.QueryRowCtx(ctx, &v, `
		SELECT id, title, description, user_id, object_key, cover_url, category,
			duration, file_size, file_hash, total_chunks, upload_id, status, play_count, like_count, created_at, updated_at, deleted_at
		FROM videos WHERE upload_id = ? AND deleted_at IS NULL
	`, uploadId)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
