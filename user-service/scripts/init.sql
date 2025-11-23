-- 用户服务数据库初始化脚本

USE user_service;

-- 创建用户表
CREATE TABLE IF NOT EXISTS `user` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `user_uuid` VARCHAR(36) NOT NULL COMMENT '用户UUID',
    `account` VARCHAR(50) NOT NULL COMMENT '用户账号',
    `password` VARCHAR(255) NOT NULL COMMENT '用户密码（加密后）',
    `avatar_url` VARCHAR(512) DEFAULT '' COMMENT '用户头像URL',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted` TINYINT UNSIGNED DEFAULT 0 COMMENT '是否删除：0-未删除，1-已删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_uuid` (`user_uuid`),
    UNIQUE KEY `uk_account` (`account`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 插入测试数据
INSERT INTO `user` (`user_uuid`, `account`, `password`) VALUES 
('550e8400-e29b-41d4-a716-446655440000', 'testuser', '$2a$10$N9qo8uLOickgx2ZMRZoMye7I6ZQ7hD13wK1Y9/1p92ledvHSKlSaa'), -- 密码: secret
('550e8400-e29b-41d4-a716-446655440001', 'admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi.'); -- 密码: password

-- 显示表结构
DESCRIBE `user`;

-- 显示插入的数据
SELECT * FROM `user`;