-- 视频服务数据库初始化脚本

CREATE DATABASE IF NOT EXISTS video_service;
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
    UNIQUE KEY uk_video_user (video_uuid, user_uuid),
    INDEX idx_video_uuid (video_uuid),
    INDEX idx_user_uuid (user_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入测试数据
INSERT INTO videos (uuid, user_uuid, title, description, duration, resolution, format, file_size, status, view_count, like_count) VALUES
('video-uuid-1', '550e8400-e29b-41d4-a716-446655440000', '测试视频1', '这是一个测试视频', 120, '1920x1080', 'mp4', 13780898, 'published', 100, 10),
('video-uuid-2', '550e8400-e29b-41d4-a716-446655440001', '测试视频2', '这是另一个测试视频', 180, '1280x720', 'mp4', 8900000, 'published', 50, 5);