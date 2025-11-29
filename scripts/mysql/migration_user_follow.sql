-- ============================================
-- Go Video Platform - 用户关注系统数据库迁移
-- 版本: v1.1.0
-- 日期: 2025-11-29
-- ============================================

USE user_service;

-- 1. 为 user 表添加新字段
-- ============================================
ALTER TABLE `user`
ADD COLUMN `nickname` VARCHAR(64) DEFAULT '' COMMENT '用户昵称' AFTER `account`,
ADD COLUMN `description` VARCHAR(512) DEFAULT '' COMMENT '个人简介' AFTER `avatar_url`,
ADD COLUMN `cover_url` VARCHAR(512) DEFAULT '' COMMENT '用户主页封面图URL' AFTER `description`;

-- 可选：为昵称添加索引（如果需要按昵称搜索）
-- ALTER TABLE `user` ADD INDEX `idx_nickname` (`nickname`);


-- 2. 创建用户关注关系表
-- ============================================
CREATE TABLE IF NOT EXISTS `user_follow` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_uuid` VARCHAR(64) NOT NULL COMMENT '关注者用户UUID',
  `target_uuid` VARCHAR(64) NOT NULL COMMENT '被关注者用户UUID',
  `status` VARCHAR(32) DEFAULT 'Following' COMMENT '关注状态: Following, Unfollowed',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（关注时间）',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '删除时间（软删除，用于取消关注）',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_target` (`user_uuid`, `target_uuid`),
  KEY `idx_user_uuid` (`user_uuid`),
  KEY `idx_target_uuid` (`target_uuid`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关注关系表';


-- 3. 验证表结构
-- ============================================

-- 查看 user 表结构
SHOW COLUMNS FROM `user`;

-- 查看 user_follow 表结构  
SHOW COLUMNS FROM `user_follow`;

-- 显示完成信息
SELECT '✅ 用户关注系统数据库迁移完成！' AS Status;

-- ============================================
-- 使用说明
-- ============================================
-- 1. 执行此脚本：
--    mysql -h 127.0.0.1 -P 3306 -u root -p < scripts/mysql/migration_user_follow.sql
--
-- 2. 字段说明：
--    - user.nickname: 用户昵称（展示用）
--    - user.description: 个人简介
--    - user.cover_url: 用户主页封面图
--    - user_follow: 关注关系表（支持软删除）
--
-- 3. 索引设计：
--    - uk_user_target: 防止重复关注
--    - idx_user_uuid: 查询某用户的关注列表
--    - idx_target_uuid: 查询某用户的粉丝列表
--    - idx_deleted_at: 软删除查询优化
-- ============================================
