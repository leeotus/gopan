// package model 定义视频相关的数据库表映射结构体。
package model

import (
	"database/sql"
	"time"
)

// Video 映射 videos 表，存储用户上传的视频元数据。
// Status 生命周期: 0(上传中) → 1(转码中) → 2(正常) 或 3(审核中) → 4(下架)
type Video struct {
	Id          int64        `db:"id"`          // 主键，自增
	Title       string       `db:"title"`       // 视频标题，最大 512 字符（约 170 个中文字）
	Description string       `db:"description"`  // 视频简介
	UserId      int64        `db:"user_id"`      // 上传者用户 ID
	ObjectKey   string       `db:"object_key"`   // 在 MinIO 中的存储 key
	CoverUrl    string       `db:"cover_url"`    // 封面图片 URL
	Category    string       `db:"category"`     // 分类标签
	Duration    int32        `db:"duration"`     // 视频时长，单位秒
	FileSize    int64        `db:"file_size"`    // 原始文件大小，单位字节
	FileHash    string       `db:"file_hash"`    // 文件哈希值，用于秒传去重
	Status      int32        `db:"status"`       // 0:上传中 1:转码中 2:正常 3:审核中 4:下架
	PlayCount   int64        `db:"play_count"`   // 累计播放数
	LikeCount   int64        `db:"like_count"`   // 累计点赞数
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`   // 软删除，通过 UPDATE deleted_at 实现
}

// Transcode 映射 transcodes 表，存储每个视频各分辨率下的 HLS 流信息。
// 一个视频可以有多条记录（360p/480p/720p/1080p）。
type Transcode struct {
	Id         int64  `db:"id"`         // 主键
	VideoId    int64  `db:"video_id"`   // 关联的 video.id
	Resolution string `db:"resolution"` // 分辨率: 360p/480p/720p/1080p
	M3U8Url    string `db:"m3u8_url"`   // HLS m3u8 播放列表地址
	Bitrate    int32  `db:"bitrate"`    // 码率，单位 kbps
}
