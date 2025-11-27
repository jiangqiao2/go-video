DROP DATABASE IF EXISTS user_service;
DROP DATABASE IF EXISTS upload_service;
DROP DATABASE IF EXISTS transcode_service;
DROP DATABASE IF EXISTS video_service;
CREATE DATABASE IF NOT EXISTS user_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE user_service;
CREATE TABLE IF NOT EXISTS user (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_uuid VARCHAR(36) NOT NULL,
  account VARCHAR(50) NOT NULL,
  password VARCHAR(255) NOT NULL,
  avatar_url VARCHAR(512),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_user_uuid (user_uuid),
  UNIQUE KEY uk_account (account)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_follow (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_uuid VARCHAR(36) NOT NULL COMMENT '关注者',
  target_uuid VARCHAR(36) NOT NULL COMMENT '被关注者',
  status VARCHAR(20) DEFAULT 'Following',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  UNIQUE KEY uk_user_target (user_uuid, target_uuid),
  INDEX idx_target (target_uuid),
  INDEX idx_user_uuid (user_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE DATABASE IF NOT EXISTS upload_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE upload_service;
CREATE TABLE IF NOT EXISTS upload_video (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  upload_video_uuid VARCHAR(36) NOT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  file_size BIGINT NOT NULL,
  file_hash VARCHAR(64),
  total_chunks INT NOT NULL,
  uploaded_chunks INT NOT NULL DEFAULT 0,
  chunk_storage_path VARCHAR(512),
  status VARCHAR(20),
  storage_path VARCHAR(512),
  completed_time TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_upload_video_uuid (upload_video_uuid),
  INDEX idx_user_uuid (user_uuid),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS upload_chunk (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  chunk_uuid VARCHAR(36) NOT NULL,
  upload_video_uuid VARCHAR(36) NOT NULL,
  chunk_index INT NOT NULL,
  chunk_hash VARCHAR(64),
  chunk_size BIGINT NOT NULL,
  storage_path VARCHAR(512),
  put_url VARCHAR(1000),
  presign_expired_at TIMESTAMP NULL,
  status VARCHAR(20),
  completed_time TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_chunk_uuid (chunk_uuid),
  UNIQUE KEY uk_upload_chunk (upload_video_uuid, chunk_index),
  INDEX idx_upload_video_uuid (upload_video_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS video_publish (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  video_uuid VARCHAR(36) NOT NULL,
  upload_video_uuid VARCHAR(36) NOT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  title VARCHAR(200) NOT NULL,
  description TEXT,
  cover_url VARCHAR(512),
  tags_json TEXT,
  status VARCHAR(20) NOT NULL DEFAULT 'Draft',
  published_at TIMESTAMP NULL,
  transcode_task_uuid VARCHAR(36),
  video_url VARCHAR(512),
  error_message VARCHAR(500),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_video_uuid (video_uuid),
  INDEX idx_upload_video_uuid (upload_video_uuid),
  INDEX idx_user_uuid (user_uuid),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS tag (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  tag_uuid VARCHAR(36) NOT NULL,
  name VARCHAR(64) NOT NULL,
  code VARCHAR(64) NOT NULL,
  description VARCHAR(256),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_tag_uuid (tag_uuid),
  UNIQUE KEY uk_tag_name (name),
  UNIQUE KEY uk_tag_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS video_tag (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  video_uuid VARCHAR(36) NOT NULL,
  tag_uuid VARCHAR(36) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_video_tag (video_uuid, tag_uuid),
  INDEX idx_video (video_uuid),
  INDEX idx_tag (tag_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE DATABASE IF NOT EXISTS transcode_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE transcode_service;
CREATE TABLE IF NOT EXISTS transcode_tasks (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  task_uuid VARCHAR(36) NOT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  video_uuid VARCHAR(36) NOT NULL,
  input_path VARCHAR(512) NOT NULL,
  output_path VARCHAR(512) NOT NULL,
  resolution VARCHAR(50) NOT NULL,
  bitrate VARCHAR(50) NOT NULL,
  status VARCHAR(20) NOT NULL,
  progress INT NOT NULL DEFAULT 0,
  message VARCHAR(255) DEFAULT '',
  hls_enabled TINYINT NOT NULL DEFAULT 0,
  hls_status VARCHAR(20),
  hls_progress INT DEFAULT 0,
  hls_output_path VARCHAR(512),
  hls_segment_duration INT DEFAULT 10,
  hls_list_size INT DEFAULT 0,
  hls_format VARCHAR(20) DEFAULT 'ts',
  hls_error_message VARCHAR(500),
  hls_started_at TIMESTAMP NULL,
  hls_completed_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_task_uuid (task_uuid),
  INDEX idx_user_uuid (user_uuid),
  INDEX idx_video_uuid (video_uuid),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS transcode_jobs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  job_uuid VARCHAR(36) NOT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  video_uuid VARCHAR(36) NOT NULL,
  input_path VARCHAR(512) NOT NULL,
  output_path VARCHAR(512) NOT NULL,
  resolution VARCHAR(50) NOT NULL,
  bitrate VARCHAR(50) NOT NULL,
  status VARCHAR(20) NOT NULL,
  progress INT DEFAULT 0,
  message VARCHAR(255),
  worker_id VARCHAR(36),
  priority INT DEFAULT 5,
  retry_count INT DEFAULT 0,
  max_retry_count INT DEFAULT 3,
  started_at TIMESTAMP NULL,
  completed_at TIMESTAMP NULL,
  estimated_time BIGINT NULL,
  actual_time BIGINT NULL,
  metadata JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_job_uuid (job_uuid),
  INDEX idx_user_uuid (user_uuid),
  INDEX idx_video_uuid (video_uuid),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS hls_jobs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  job_uuid VARCHAR(36) NOT NULL,
  user_uuid VARCHAR(36) NOT NULL,
  video_uuid VARCHAR(36) NOT NULL,
  source_job_uuid VARCHAR(36),
  source_type VARCHAR(20),
  input_path VARCHAR(512) NOT NULL,
  output_dir VARCHAR(512) NOT NULL,
  master_playlist VARCHAR(512),
  profiles_json JSON,
  status VARCHAR(20) NOT NULL,
  progress INT DEFAULT 0,
  segment_duration INT DEFAULT 10,
  list_size INT DEFAULT 0,
  format VARCHAR(20) DEFAULT 'mpegts',
  variant_count INT DEFAULT 0,
  error_message VARCHAR(500),
  started_at TIMESTAMP NULL,
  completed_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT DEFAULT 0,
  UNIQUE KEY uk_hls_job_uuid (job_uuid),
  INDEX idx_user_uuid (user_uuid),
  INDEX idx_video_uuid (video_uuid),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
