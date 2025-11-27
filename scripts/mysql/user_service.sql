-- 用户服务数据库初始化脚本

CREATE DATABASE IF NOT EXISTS user_service;
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

CREATE TABLE IF NOT EXISTS user_follow (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_uuid VARCHAR(36) NOT NULL,
    target_uuid VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'Following',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY uk_user_target (user_uuid, target_uuid),
    INDEX idx_target (target_uuid),
    INDEX idx_user_uuid (user_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入测试数据
INSERT INTO users (uuid, username, email, password_hash, nickname, status) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'admin', 'admin@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iYqiSuAZDOyy.JjPGw0jUBKKNjO6', '管理员', 1),
('550e8400-e29b-41d4-a716-446655440001', 'testuser', 'test@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iYqiSuAZDOyy.JjPGw0jUBKKNjO6', '测试用户', 1);
