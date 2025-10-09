-- 上传服务数据库初始化脚本

CREATE DATABASE IF NOT EXISTS upload_service;
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

-- 上传视频表（基于DDD架构的视频上传管理）
CREATE TABLE IF NOT EXISTS upload_video (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    upload_video_uuid VARCHAR(36) UNIQUE NOT NULL COMMENT '上传视频唯一UUID',
    user_uuid VARCHAR(36) NOT NULL COMMENT '上传用户的唯一UUID',
    file_name VARCHAR(255) NOT NULL COMMENT '原始文件名',
    file_size INT NOT NULL COMMENT '上传文件大小',
    file_hash VARCHAR(64) NOT NULL COMMENT '文件内容Hash',
    total_chunks INT NOT NULL COMMENT '分片数量',
    uploaded_chunks INT DEFAULT 0 COMMENT '已经完成的分片数量',
    chunk_storage_path VARCHAR(500) COMMENT '分片路径',
    status VARCHAR(50) NOT NULL COMMENT '状态',
    storage_path VARCHAR(500) COMMENT '合并后在Minio的路径',
    completed_time TIMESTAMP NULL COMMENT '完成时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_deleted BIGINT DEFAULT 0 COMMENT '软删除标记',
    INDEX idx_upload_video_uuid (upload_video_uuid),
    INDEX idx_user_uuid (user_uuid),
    INDEX idx_file_hash (file_hash),
    INDEX idx_status (status),
    INDEX idx_is_deleted (is_deleted),
    UNIQUE KEY uk_user_file (user_uuid, file_name, file_hash, file_size, is_deleted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上传视频表';

-- 上传分片表（基于DDD架构的分片管理）
CREATE TABLE IF NOT EXISTS upload_chunk (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    chunk_uuid VARCHAR(36) UNIQUE NOT NULL COMMENT '分片唯一UUID',
    upload_video_uuid VARCHAR(36) NOT NULL COMMENT '关联上传视频的uuid',
    chunk_index INT NOT NULL COMMENT '分片索引，从0开始',
    chunk_hash VARCHAR(64) COMMENT '分片Hash',
    chunk_size INT NOT NULL COMMENT '分片大小',
    storage_path VARCHAR(500) COMMENT '在Minio的存储路径',
    status VARCHAR(50) NOT NULL COMMENT '状态',
    completed_time TIMESTAMP NULL COMMENT '完成时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_deleted BIGINT DEFAULT 0 COMMENT '软删除标记',
    INDEX idx_chunk_uuid (chunk_uuid),
    INDEX idx_upload_video_uuid (upload_video_uuid),
    INDEX idx_chunk_index (chunk_index),
    INDEX idx_status (status),
    INDEX idx_is_deleted (is_deleted),
    UNIQUE KEY uk_video_chunk (upload_video_uuid, chunk_index, is_deleted),
    FOREIGN KEY (upload_video_uuid) REFERENCES upload_video(upload_video_uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上传分片表';