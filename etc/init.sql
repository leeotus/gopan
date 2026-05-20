-- GoPan VOD Platform - 数据库初始化脚本

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NOT NULL DEFAULT '',
    `password` VARCHAR(128) NOT NULL DEFAULT '',
    `email` VARCHAR(128) NOT NULL DEFAULT '',
    `avatar` VARCHAR(512) NOT NULL DEFAULT '',
    `signature` VARCHAR(255) NOT NULL DEFAULT '',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 视频表
CREATE TABLE IF NOT EXISTS `videos` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `title` VARCHAR(255) NOT NULL DEFAULT '',
    `description` TEXT NOT NULL,
    `user_id` BIGINT NOT NULL DEFAULT 0,
    `object_key` VARCHAR(512) NOT NULL DEFAULT '',
    `cover_url` VARCHAR(512) NOT NULL DEFAULT '',
    `category` VARCHAR(64) NOT NULL DEFAULT '',
    `duration` INT NOT NULL DEFAULT 0 COMMENT '时长(秒)',
    `file_size` BIGINT NOT NULL DEFAULT 0,
    `file_hash` VARCHAR(64) NOT NULL DEFAULT '',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0:上传中 1:转码中 2:正常 3:审核中 4:下架',
    `play_count` BIGINT NOT NULL DEFAULT 0,
    `like_count` BIGINT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_category` (`category`),
    KEY `idx_status` (`status`),
    KEY `idx_play_count` (`play_count`),
    KEY `idx_file_hash` (`file_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 转码信息表
CREATE TABLE IF NOT EXISTS `transcodes` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `video_id` BIGINT NOT NULL DEFAULT 0,
    `resolution` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '360p/480p/720p/1080p',
    `m3u8_url` VARCHAR(512) NOT NULL DEFAULT '',
    `bitrate` INT NOT NULL DEFAULT 0 COMMENT 'kbps',
    PRIMARY KEY (`id`),
    KEY `idx_video_id` (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 点赞表
CREATE TABLE IF NOT EXISTS `likes` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL DEFAULT 0,
    `video_id` BIGINT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_video` (`user_id`, `video_id`),
    KEY `idx_video_id` (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 收藏表
CREATE TABLE IF NOT EXISTS `favorites` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL DEFAULT 0,
    `video_id` BIGINT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_video` (`user_id`, `video_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 评论表
CREATE TABLE IF NOT EXISTS `comments` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL DEFAULT 0,
    `video_id` BIGINT NOT NULL DEFAULT 0,
    `parent_id` BIGINT NOT NULL DEFAULT 0 COMMENT '父评论ID，0为顶级评论',
    `content` TEXT NOT NULL,
    `like_count` INT NOT NULL DEFAULT 0,
    `reply_count` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_video_id_parent` (`video_id`, `parent_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 弹幕表
CREATE TABLE IF NOT EXISTS `danmakus` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL DEFAULT 0,
    `video_id` BIGINT NOT NULL DEFAULT 0,
    `content` VARCHAR(255) NOT NULL DEFAULT '',
    `time` DOUBLE NOT NULL DEFAULT 0 COMMENT '弹幕出现时间(秒)',
    `color` VARCHAR(16) NOT NULL DEFAULT '#ffffff',
    `mode` TINYINT NOT NULL DEFAULT 1 COMMENT '1:滚动 2:顶部 3:底部',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_video_id_time` (`video_id`, `time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
