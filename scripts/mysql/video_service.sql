-- 视频服务数据库初始化脚本

CREATE DATABASE IF NOT EXISTS video_service;
USE video_service;

CREATE TABLE IF NOT EXISTS video (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  video_uuid VARCHAR(36) NOT NULL,
  upload_video_uuid VARCHAR(36) DEFAULT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  cover_url VARCHAR(512),
  video_url VARCHAR(512),
  tags_json TEXT,
  status VARCHAR(20) DEFAULT 'Published',
  like_count BIGINT DEFAULT 0,
  play_count BIGINT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_video_uuid (video_uuid),
  INDEX idx_user_uuid (user_uuid),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS video_like (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_uuid VARCHAR(36) NOT NULL,
  video_uuid VARCHAR(36) NOT NULL,
  status VARCHAR(20) DEFAULT 'Liked',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_video (user_uuid, video_uuid),
  INDEX idx_video (video_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS video_comment (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  comment_uuid VARCHAR(36) NOT NULL,
  video_uuid VARCHAR(36) NOT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  content TEXT NOT NULL,
  parent_uuid VARCHAR(36) DEFAULT NULL,
  like_count BIGINT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  UNIQUE KEY uk_comment_uuid (comment_uuid),
  INDEX idx_video_uuid (video_uuid),
  INDEX idx_parent_uuid (parent_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
