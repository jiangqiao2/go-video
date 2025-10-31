-- 转码服务数据库初始化脚本

-- 创建数据库
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除标记',
    
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_user_uuid (user_uuid),
    INDEX idx_video_uuid (video_uuid),
    INDEX idx_status (status),
    INDEX idx_worker_id (worker_id),
    INDEX idx_priority (priority),
    INDEX idx_created_at (created_at)
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
    last_heartbeat_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '最后心跳时间',
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    system_info JSON COMMENT '系统信息',
    metadata JSON COMMENT '元数据',
    
    INDEX idx_worker_id (worker_id),
    INDEX idx_status (status),
    INDEX idx_last_heartbeat (last_heartbeat_at),
    INDEX idx_status_heartbeat (status, last_heartbeat_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Worker表';

-- 创建任务执行日志表（可选）
CREATE TABLE IF NOT EXISTS task_execution_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    task_id VARCHAR(36) NOT NULL COMMENT '任务ID',
    worker_id VARCHAR(36) COMMENT 'Worker ID',
    log_level VARCHAR(10) NOT NULL COMMENT '日志级别',
    message TEXT NOT NULL COMMENT '日志消息',
    details JSON COMMENT '详细信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_task_id (task_id),
    INDEX idx_worker_id (worker_id),
    INDEX idx_created_at (created_at),
    INDEX idx_task_created (task_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务执行日志表';

-- 创建系统配置表（可选）
CREATE TABLE IF NOT EXISTS system_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    config_key VARCHAR(100) NOT NULL UNIQUE COMMENT '配置键',
    config_value TEXT NOT NULL COMMENT '配置值',
    description VARCHAR(255) COMMENT '配置描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 插入默认配置
INSERT INTO system_configs (config_key, config_value, description) VALUES
('max_concurrent_tasks_per_worker', '4', 'Worker最大并发任务数'),
('task_timeout_seconds', '3600', '任务超时时间（秒）'),
('worker_heartbeat_timeout_seconds', '60', 'Worker心跳超时时间（秒）'),
('max_retry_count', '3', '任务最大重试次数'),
('cleanup_interval_minutes', '30', '清理任务间隔（分钟）')
ON DUPLICATE KEY UPDATE 
    config_value = VALUES(config_value),
    updated_at = CURRENT_TIMESTAMP;

-- 创建用户（如果需要）
-- CREATE USER IF NOT EXISTS 'transcode_user'@'%' IDENTIFIED BY 'transcode_password';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON transcode_service.* TO 'transcode_user'@'%';
-- FLUSH PRIVILEGES;

-- 创建视图：任务统计
CREATE OR REPLACE VIEW task_statistics AS
SELECT 
    COUNT(*) as total_tasks,
    SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
    SUM(CASE WHEN status = 'assigned' THEN 1 ELSE 0 END) as assigned_tasks,
    SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END) as processing_tasks,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_tasks,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_tasks,
    SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled_tasks,
    SUM(CASE WHEN status = 'retrying' THEN 1 ELSE 0 END) as retrying_tasks,
    AVG(CASE WHEN status = 'completed' AND actual_time IS NOT NULL THEN actual_time/1000000000 ELSE NULL END) as avg_completion_time_seconds
FROM transcode_tasks;

-- 创建视图：Worker统计
CREATE OR REPLACE VIEW worker_statistics AS
SELECT 
    COUNT(*) as total_workers,
    SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) as online_workers,
    SUM(CASE WHEN status = 'offline' THEN 1 ELSE 0 END) as offline_workers,
    SUM(CASE WHEN status = 'busy' THEN 1 ELSE 0 END) as busy_workers,
    SUM(CASE WHEN status = 'idle' THEN 1 ELSE 0 END) as idle_workers,
    SUM(CASE WHEN status = 'maintenance' THEN 1 ELSE 0 END) as maintenance_workers,
    SUM(current_tasks) as total_running_tasks,
    SUM(max_tasks) as total_capacity,
    AVG(cpu_usage) as avg_cpu_usage,
    AVG(memory_usage) as avg_memory_usage,
    AVG(CASE WHEN max_tasks > 0 THEN current_tasks * 100.0 / max_tasks ELSE 0 END) as avg_load_percentage
FROM workers;

-- 创建存储过程：清理过期任务
DELIMITER //
CREATE PROCEDURE CleanupExpiredTasks(IN hours_threshold INT)
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE task_count INT DEFAULT 0;
    
    -- 更新超过指定小时数且未完成的任务为已取消
    UPDATE transcode_tasks 
    SET status = 'cancelled', 
        updated_at = CURRENT_TIMESTAMP,
        message = CONCAT('Task expired after ', hours_threshold, ' hours')
    WHERE status NOT IN ('completed', 'failed', 'cancelled')
      AND created_at < DATE_SUB(NOW(), INTERVAL hours_threshold HOUR);
    
    -- 获取受影响的行数
    SET task_count = ROW_COUNT();
    
    -- 返回清理的任务数量
    SELECT task_count as cleaned_tasks;
END //
DELIMITER ;

-- 创建存储过程：清理不健康的Worker
DELIMITER //
CREATE PROCEDURE CleanupUnhealthyWorkers(IN timeout_minutes INT)
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE worker_count INT DEFAULT 0;
    
    -- 更新超过指定分钟数未发送心跳的Worker为离线
    UPDATE workers 
    SET status = 'offline', 
        current_tasks = 0,
        updated_at = CURRENT_TIMESTAMP
    WHERE status NOT IN ('offline', 'maintenance')
      AND last_heartbeat_at < DATE_SUB(NOW(), INTERVAL timeout_minutes MINUTE);
    
    -- 获取受影响的行数
    SET worker_count = ROW_COUNT();
    
    -- 将这些Worker的任务重新设置为待处理
    UPDATE transcode_tasks 
    SET status = 'pending', 
        worker_id = NULL,
        updated_at = CURRENT_TIMESTAMP
    WHERE status IN ('assigned', 'processing')
      AND worker_id IN (
          SELECT worker_id FROM workers 
          WHERE status = 'offline' 
            AND last_heartbeat_at < DATE_SUB(NOW(), INTERVAL timeout_minutes MINUTE)
      );
    
    -- 返回清理的Worker数量
    SELECT worker_count as cleaned_workers;
END //
DELIMITER ;

-- 创建事件调度器（如果MySQL支持）
-- SET GLOBAL event_scheduler = ON;
-- 
-- CREATE EVENT IF NOT EXISTS cleanup_expired_tasks
-- ON SCHEDULE EVERY 1 HOUR
-- DO
--   CALL CleanupExpiredTasks(24);
-- 
-- CREATE EVENT IF NOT EXISTS cleanup_unhealthy_workers
-- ON SCHEDULE EVERY 5 MINUTE
-- DO
--   CALL CleanupUnhealthyWorkers(5);

COMMIT;
