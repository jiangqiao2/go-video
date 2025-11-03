-- 视频上传系统数据库初始化脚本
-- 该脚本将初始化所有微服务的数据库和表结构

-- ========================================
-- 用户服务数据库初始化
-- ========================================

CREATE DATABASE IF NOT EXISTS user_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE user_service;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar_url VARCHAR(500),
    status TINYINT DEFAULT 1 COMMENT '1:正常 2:待激活 3:已禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_uuid (uuid),
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入测试数据
INSERT INTO users (uuid, username, email, password_hash, nickname, status) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'admin', 'admin@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iYqiSuAZDOyy.JjPGw0jUBKKNjO6', '管理员', 1),
('550e8400-e29b-41d4-a716-446655440001', 'testuser', 'test@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iYqiSuAZDOyy.JjPGw0jUBKKNjO6', '测试用户', 1)
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;

-- ========================================
-- 上传服务数据库初始化
-- ========================================

CREATE DATABASE IF NOT EXISTS upload_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE upload_service;

-- 上传任务表
CREATE TABLE IF NOT EXISTS upload_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    user_uuid VARCHAR(36) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100),
    storage_path VARCHAR(500),
    status ENUM('pending', 'uploading', 'completed', 'failed') DEFAULT 'pending',
    progress INT DEFAULT 0,
    error_msg TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_uuid (uuid),
    INDEX idx_user_uuid (user_uuid),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 上传分片表（支持大文件分片上传）
CREATE TABLE IF NOT EXISTS upload_chunks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_uuid VARCHAR(36) NOT NULL,
    chunk_index INT NOT NULL,
    chunk_size BIGINT NOT NULL,
    chunk_hash VARCHAR(64),
    status ENUM('pending', 'completed', 'failed') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_chunk_index (chunk_index),
    UNIQUE KEY uk_task_chunk (task_uuid, chunk_index)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 视频发布元数据表
CREATE TABLE IF NOT EXISTS video_publish (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_uuid VARCHAR(36) NOT NULL,
    upload_video_uuid VARCHAR(36) NOT NULL,
    user_uuid VARCHAR(36) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    cover_url VARCHAR(512),
    tags_json TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'Draft',
    published_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_deleted TINYINT DEFAULT 0,
    UNIQUE KEY uk_video_uuid (video_uuid),
    UNIQUE KEY uk_upload_video_uuid (upload_video_uuid),
    INDEX idx_user_uuid (user_uuid),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ========================================
-- 视频服务数据库初始化
-- ========================================

CREATE DATABASE IF NOT EXISTS video_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE video_service;

-- 视频元数据表
CREATE TABLE IF NOT EXISTS videos (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    user_uuid VARCHAR(36) NOT NULL,
    upload_task_uuid VARCHAR(36), -- 关联上传任务
    title VARCHAR(255) NOT NULL,
    description TEXT,
    duration INT, -- 视频时长（秒）
    resolution VARCHAR(20), -- 分辨率
    format VARCHAR(20),
    file_size BIGINT,
    thumbnail_url VARCHAR(500), -- 缩略图
    video_url VARCHAR(500), -- 播放地址
    status ENUM('processing', 'published', 'deleted') DEFAULT 'processing',
    view_count BIGINT DEFAULT 0,
    like_count BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_uuid (uuid),
    INDEX idx_user_uuid (user_uuid),
    INDEX idx_upload_task_uuid (upload_task_uuid),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_view_count (view_count),
    INDEX idx_like_count (like_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 视频播放记录表
CREATE TABLE IF NOT EXISTS video_play_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_uuid VARCHAR(36) NOT NULL,
    user_uuid VARCHAR(36),
    play_duration INT, -- 播放时长
    play_progress DECIMAL(5,2), -- 播放进度百分比
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_video_uuid (video_uuid),
    INDEX idx_user_uuid (user_uuid),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 视频点赞表
CREATE TABLE IF NOT EXISTS video_likes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_uuid VARCHAR(36) NOT NULL,
    user_uuid VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_video_uuid (video_uuid),
    INDEX idx_user_uuid (user_uuid),
    UNIQUE KEY uk_video_user (video_uuid, user_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ========================================
-- 转码服务数据库初始化
-- ========================================

CREATE DATABASE IF NOT EXISTS transcode_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE transcode_service;

CREATE TABLE IF NOT EXISTS transcode_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_uuid VARCHAR(36) NOT NULL UNIQUE COMMENT '任务UUID',
    user_uuid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    video_uuid VARCHAR(36) NOT NULL COMMENT '关联视频UUID',
    input_path VARCHAR(512) NOT NULL COMMENT '输入视频路径',
    output_path VARCHAR(512) NOT NULL COMMENT '输出路径',
    resolution VARCHAR(50) NOT NULL COMMENT '转码分辨率',
    bitrate VARCHAR(50) NOT NULL COMMENT '转码码率',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '任务状态',
    progress INT NOT NULL DEFAULT 0 COMMENT '转码进度(0-100)',
    message VARCHAR(255) DEFAULT '' COMMENT '状态描述或错误信息',
    worker_id VARCHAR(36) DEFAULT NULL COMMENT '分配的Worker ID',
    priority INT NOT NULL DEFAULT 5 COMMENT '任务优先级(1-10)',
    retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    max_retry_count INT NOT NULL DEFAULT 3 COMMENT '最大重试次数',
    started_at TIMESTAMP NULL COMMENT '开始时间',
    completed_at TIMESTAMP NULL COMMENT '完成时间',
    estimated_time BIGINT NULL COMMENT '预估耗时(纳秒)',
    actual_time BIGINT NULL COMMENT '实际耗时(纳秒)',
    metadata JSON COMMENT '元数据',
    -- HLS相关字段
    hls_enabled TINYINT NOT NULL DEFAULT 0 COMMENT 'HLS切片是否启用',
    hls_status VARCHAR(20) DEFAULT NULL COMMENT 'HLS切片状态(pending/processing/completed/failed)',
    hls_progress INT DEFAULT 0 COMMENT 'HLS切片进度(0-100)',
    hls_output_path VARCHAR(512) DEFAULT NULL COMMENT 'HLS输出路径',
    hls_segment_duration INT DEFAULT 10 COMMENT 'HLS切片时长(秒)',
    hls_list_size INT DEFAULT 0 COMMENT 'HLS播放列表大小(0表示无限制)',
    hls_format VARCHAR(20) DEFAULT 'ts' COMMENT 'HLS切片格式(ts/fmp4)',
    hls_error_message VARCHAR(500) DEFAULT NULL COMMENT 'HLS切片错误信息',
    hls_started_at TIMESTAMP NULL COMMENT 'HLS切片开始时间',
    hls_completed_at TIMESTAMP NULL COMMENT 'HLS切片完成时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除标记',
    
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_user_uuid (user_uuid),
    INDEX idx_video_uuid (video_uuid),
    INDEX idx_status (status),
    INDEX idx_worker_id (worker_id),
    INDEX idx_priority (priority),
    INDEX idx_created_at (created_at),
    INDEX idx_hls_enabled (hls_enabled),
    INDEX idx_hls_status (hls_status),
    INDEX idx_hls_enabled_status (hls_enabled, hls_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='转码任务表';

-- 创建Worker表
CREATE TABLE IF NOT EXISTS workers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    worker_id VARCHAR(36) NOT NULL UNIQUE COMMENT 'Worker UUID',
    name VARCHAR(100) NOT NULL COMMENT 'Worker名称',
    status VARCHAR(20) NOT NULL DEFAULT 'offline' COMMENT 'Worker状态',
    max_tasks INT NOT NULL COMMENT '最大并发任务数',
    current_tasks INT NOT NULL DEFAULT 0 COMMENT '当前任务数',
    cpu_usage DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'CPU使用率',
    memory_usage DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '内存使用率',
    last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '最后心跳时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_worker_id (worker_id),
    INDEX idx_status (status),
    INDEX idx_last_heartbeat (last_heartbeat)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Worker表';

-- 创建任务队列表
CREATE TABLE IF NOT EXISTS task_queue (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    task_id VARCHAR(36) NOT NULL COMMENT '任务ID',
    priority INT NOT NULL DEFAULT 5 COMMENT '优先级',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_task_id (task_id),
    INDEX idx_priority_created (priority DESC, created_at ASC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务队列表';

-- 任务统计视图
CREATE OR REPLACE VIEW task_statistics AS
SELECT 
    status,
    COUNT(*) as count,
    AVG(progress) as avg_progress,
    AVG(actual_time) as avg_time,
    -- HLS统计信息
    SUM(CASE WHEN hls_enabled = 1 THEN 1 ELSE 0 END) as hls_enabled_count,
    SUM(CASE WHEN hls_status = 'completed' THEN 1 ELSE 0 END) as hls_completed_count,
    AVG(CASE WHEN hls_enabled = 1 THEN hls_progress ELSE NULL END) as avg_hls_progress
FROM transcode_tasks 
WHERE is_deleted = 0 
GROUP BY status;

-- HLS分辨率配置表
CREATE TABLE IF NOT EXISTS hls_resolution_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_uuid VARCHAR(36) NOT NULL COMMENT '关联任务UUID',
    resolution VARCHAR(50) NOT NULL COMMENT '分辨率(如720p, 1080p)',
    width INT NOT NULL COMMENT '视频宽度',
    height INT NOT NULL COMMENT '视频高度',
    bitrate VARCHAR(50) NOT NULL COMMENT '码率',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态(pending/processing/completed/failed)',
    output_path VARCHAR(512) DEFAULT NULL COMMENT '输出路径',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_resolution (resolution),
    INDEX idx_status (status),
    FOREIGN KEY (task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HLS分辨率配置表';

-- HLS切片表
CREATE TABLE IF NOT EXISTS hls_segments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_uuid VARCHAR(36) NOT NULL COMMENT '关联任务UUID',
    resolution VARCHAR(50) NOT NULL COMMENT '分辨率',
    segment_index INT NOT NULL COMMENT '切片索引',
    segment_path VARCHAR(512) NOT NULL COMMENT '切片文件路径',
    duration DECIMAL(10,6) NOT NULL COMMENT '切片时长(秒)',
    file_size BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_resolution (resolution),
    INDEX idx_segment_index (segment_index),
    INDEX idx_task_resolution (task_uuid, resolution),
    FOREIGN KEY (task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HLS切片表';

-- HLS播放列表表
CREATE TABLE IF NOT EXISTS hls_playlists (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_uuid VARCHAR(36) NOT NULL COMMENT '关联任务UUID',
    playlist_type VARCHAR(20) NOT NULL COMMENT '播放列表类型(master/media)',
    resolution VARCHAR(50) DEFAULT NULL COMMENT '分辨率(仅media类型)',
    playlist_path VARCHAR(512) NOT NULL COMMENT '播放列表文件路径',
    content TEXT NOT NULL COMMENT '播放列表内容',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_playlist_type (playlist_type),
    INDEX idx_resolution (resolution),
    FOREIGN KEY (task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HLS播放列表表';

-- ========================================
-- 数据库初始化完成
-- ========================================

SELECT 'Database initialization completed successfully!' as message;
