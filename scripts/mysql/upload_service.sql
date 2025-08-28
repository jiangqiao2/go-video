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