-- 视频服务数据库初始化脚本

CREATE DATABASE IF NOT EXISTS video_service;
USE video_service;

CREATE TABLE IF NOT EXISTS videos (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  video_uuid CHAR(36) NOT NULL UNIQUE,
  user_uuid CHAR(36) NOT NULL,
  upload_video_uuid CHAR(36) DEFAULT NULL,
  title VARCHAR(200) NOT NULL,
  description TEXT NULL,
  cover_url VARCHAR(500) NULL,
  video_url VARCHAR(500) NULL,
  duration_sec INT NULL,
  resolution VARCHAR(32) NULL,
  size_bytes BIGINT NULL,
  status ENUM('processing','published','failed','draft','private','deleted') NOT NULL DEFAULT 'processing',
  transcode_task_uuid CHAR(36) NULL,
  error_message VARCHAR(500) NULL,
  privacy ENUM('public','followers','private') NOT NULL DEFAULT 'public',
  published_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_user_status_created (user_uuid, status, created_at DESC),
  KEY idx_status_created (status, created_at DESC)
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

CREATE TABLE IF NOT EXISTS video_comment_root (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  root_uuid VARCHAR(36) NOT NULL,
  video_uuid VARCHAR(36) NOT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  content TEXT NOT NULL,
  like_count BIGINT DEFAULT 0,
  reply_count BIGINT DEFAULT 0,
  is_deleted TINYINT(1) NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  UNIQUE KEY uk_root_uuid (root_uuid),
  INDEX idx_video_created (video_uuid, created_at DESC),
  INDEX idx_video_hot (video_uuid, is_deleted, like_count, reply_count, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS video_comment_reply (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  comment_uuid VARCHAR(36) NOT NULL,
  root_uuid VARCHAR(36) NOT NULL,
  parent_uuid VARCHAR(36) NOT NULL,
  parent_type ENUM('root','reply') NOT NULL DEFAULT 'root',
  depth INT NOT NULL DEFAULT 1,
  path VARCHAR(512) NOT NULL,
  video_uuid VARCHAR(36) NOT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  content TEXT NOT NULL,
  like_count BIGINT DEFAULT 0,
  reply_count BIGINT DEFAULT 0,
  is_deleted TINYINT(1) NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  UNIQUE KEY uk_comment_uuid (comment_uuid),
  INDEX idx_root_created (root_uuid, created_at DESC),
  INDEX idx_root_parent (root_uuid, parent_uuid, created_at DESC),
  INDEX idx_root_path (root_uuid, path),
  INDEX idx_root_hot (root_uuid, is_deleted, like_count, reply_count, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS video_comment_like (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_uuid VARCHAR(36) NOT NULL,
  comment_uuid VARCHAR(36) NOT NULL,
  status VARCHAR(20) DEFAULT 'Liked',
  is_deleted TINYINT(1) NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_comment (user_uuid, comment_uuid),
  INDEX idx_comment (comment_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
