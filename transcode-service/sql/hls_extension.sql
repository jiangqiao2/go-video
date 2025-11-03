-- HLS切片扩展数据库表结构
-- 此脚本用于为转码服务添加HLS切片功能的数据库支持

USE transcode_service;

-- 为转码任务表添加HLS相关字段
ALTER TABLE transcode_tasks 
ADD COLUMN hls_enabled TINYINT NOT NULL DEFAULT 0 COMMENT 'HLS切片是否启用',
ADD COLUMN hls_status VARCHAR(20) DEFAULT NULL COMMENT 'HLS切片状态(pending/processing/completed/failed)',
ADD COLUMN hls_progress INT DEFAULT 0 COMMENT 'HLS切片进度(0-100)',
ADD COLUMN hls_output_path VARCHAR(512) DEFAULT NULL COMMENT 'HLS输出路径',
ADD COLUMN hls_segment_duration INT DEFAULT 10 COMMENT 'HLS切片时长(秒)',
ADD COLUMN hls_list_size INT DEFAULT 0 COMMENT 'HLS播放列表大小(0表示无限制)',
ADD COLUMN hls_format VARCHAR(20) DEFAULT 'ts' COMMENT 'HLS切片格式(ts/fmp4)',
ADD COLUMN hls_error_message VARCHAR(500) DEFAULT NULL COMMENT 'HLS切片错误信息',
ADD COLUMN hls_started_at TIMESTAMP NULL COMMENT 'HLS切片开始时间',
ADD COLUMN hls_completed_at TIMESTAMP NULL COMMENT 'HLS切片完成时间';

-- 添加HLS相关索引
ALTER TABLE transcode_tasks 
ADD INDEX idx_hls_enabled (hls_enabled),
ADD INDEX idx_hls_status (hls_status),
ADD INDEX idx_hls_enabled_status (hls_enabled, hls_status);

-- 创建HLS分辨率配置表
CREATE TABLE IF NOT EXISTS hls_resolution_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    task_uuid VARCHAR(36) NOT NULL COMMENT '关联的转码任务UUID',
    width INT NOT NULL COMMENT '视频宽度',
    height INT NOT NULL COMMENT '视频高度',
    bitrate VARCHAR(50) NOT NULL COMMENT '码率设置',
    output_path VARCHAR(512) DEFAULT NULL COMMENT '该分辨率的输出路径',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '该分辨率的处理状态',
    progress INT NOT NULL DEFAULT 0 COMMENT '该分辨率的处理进度(0-100)',
    error_message VARCHAR(500) DEFAULT NULL COMMENT '错误信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_status (status),
    INDEX idx_task_status (task_uuid, status),
    FOREIGN KEY (task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HLS分辨率配置表';

-- 创建HLS切片文件表
CREATE TABLE IF NOT EXISTS hls_segments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    task_uuid VARCHAR(36) NOT NULL COMMENT '关联的转码任务UUID',
    resolution_config_id BIGINT UNSIGNED NOT NULL COMMENT '关联的分辨率配置ID',
    segment_index INT NOT NULL COMMENT '切片序号',
    segment_filename VARCHAR(255) NOT NULL COMMENT '切片文件名',
    segment_path VARCHAR(512) NOT NULL COMMENT '切片文件路径',
    duration DECIMAL(10,6) NOT NULL COMMENT '切片时长(秒)',
    file_size BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_resolution_config (resolution_config_id),
    INDEX idx_segment_index (segment_index),
    INDEX idx_task_resolution (task_uuid, resolution_config_id),
    FOREIGN KEY (task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE,
    FOREIGN KEY (resolution_config_id) REFERENCES hls_resolution_configs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HLS切片文件表';

-- 创建HLS播放列表表
CREATE TABLE IF NOT EXISTS hls_playlists (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    task_uuid VARCHAR(36) NOT NULL COMMENT '关联的转码任务UUID',
    playlist_type VARCHAR(20) NOT NULL COMMENT '播放列表类型(master/media)',
    resolution_config_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联的分辨率配置ID(仅media类型)',
    filename VARCHAR(255) NOT NULL COMMENT '播放列表文件名',
    file_path VARCHAR(512) NOT NULL COMMENT '播放列表文件路径',
    content TEXT NOT NULL COMMENT '播放列表内容',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_task_uuid (task_uuid),
    INDEX idx_playlist_type (playlist_type),
    INDEX idx_resolution_config (resolution_config_id),
    INDEX idx_task_type (task_uuid, playlist_type),
    FOREIGN KEY (task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE,
    FOREIGN KEY (resolution_config_id) REFERENCES hls_resolution_configs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HLS播放列表表';

-- 更新任务统计视图，包含HLS信息
DROP VIEW IF EXISTS task_statistics;
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
    AVG(CASE WHEN status = 'completed' AND actual_time IS NOT NULL THEN actual_time/1000000000 ELSE NULL END) as avg_completion_time_seconds,
    -- HLS统计
    SUM(CASE WHEN hls_enabled = 1 THEN 1 ELSE 0 END) as hls_enabled_tasks,
    SUM(CASE WHEN hls_enabled = 1 AND hls_status = 'pending' THEN 1 ELSE 0 END) as hls_pending_tasks,
    SUM(CASE WHEN hls_enabled = 1 AND hls_status = 'processing' THEN 1 ELSE 0 END) as hls_processing_tasks,
    SUM(CASE WHEN hls_enabled = 1 AND hls_status = 'completed' THEN 1 ELSE 0 END) as hls_completed_tasks,
    SUM(CASE WHEN hls_enabled = 1 AND hls_status = 'failed' THEN 1 ELSE 0 END) as hls_failed_tasks
FROM transcode_tasks;

-- 创建HLS统计视图
CREATE OR REPLACE VIEW hls_statistics AS
SELECT 
    t.task_uuid,
    t.hls_status,
    t.hls_progress,
    COUNT(rc.id) as total_resolutions,
    SUM(CASE WHEN rc.status = 'completed' THEN 1 ELSE 0 END) as completed_resolutions,
    SUM(CASE WHEN rc.status = 'failed' THEN 1 ELSE 0 END) as failed_resolutions,
    COUNT(s.id) as total_segments,
    SUM(s.file_size) as total_segments_size,
    COUNT(p.id) as total_playlists
FROM transcode_tasks t
LEFT JOIN hls_resolution_configs rc ON t.task_uuid = rc.task_uuid
LEFT JOIN hls_segments s ON t.task_uuid = s.task_uuid
LEFT JOIN hls_playlists p ON t.task_uuid = p.task_uuid
WHERE t.hls_enabled = 1
GROUP BY t.task_uuid, t.hls_status, t.hls_progress;

-- 创建存储过程：清理HLS相关数据
DELIMITER //
CREATE PROCEDURE CleanupHLSData(IN task_uuid_param VARCHAR(36))
BEGIN
    DECLARE done INT DEFAULT FALSE;
    
    -- 删除HLS切片文件记录
    DELETE FROM hls_segments WHERE task_uuid = task_uuid_param;
    
    -- 删除HLS播放列表记录
    DELETE FROM hls_playlists WHERE task_uuid = task_uuid_param;
    
    -- 删除HLS分辨率配置记录
    DELETE FROM hls_resolution_configs WHERE task_uuid = task_uuid_param;
    
    -- 清理转码任务表中的HLS字段
    UPDATE transcode_tasks 
    SET hls_status = NULL,
        hls_progress = 0,
        hls_output_path = NULL,
        hls_error_message = NULL,
        hls_started_at = NULL,
        hls_completed_at = NULL
    WHERE task_uuid = task_uuid_param;
    
    SELECT 'HLS data cleaned successfully' as result;
END //
DELIMITER ;

-- 创建存储过程：获取HLS任务详情
DELIMITER //
CREATE PROCEDURE GetHLSTaskDetails(IN task_uuid_param VARCHAR(36))
BEGIN
    -- 获取任务基本信息
    SELECT 
        task_uuid,
        hls_enabled,
        hls_status,
        hls_progress,
        hls_output_path,
        hls_segment_duration,
        hls_list_size,
        hls_format,
        hls_error_message,
        hls_started_at,
        hls_completed_at
    FROM transcode_tasks 
    WHERE task_uuid = task_uuid_param;
    
    -- 获取分辨率配置信息
    SELECT 
        id,
        width,
        height,
        bitrate,
        output_path,
        status,
        progress,
        error_message
    FROM hls_resolution_configs 
    WHERE task_uuid = task_uuid_param
    ORDER BY width DESC, height DESC;
    
    -- 获取播放列表信息
    SELECT 
        playlist_type,
        filename,
        file_path,
        resolution_config_id
    FROM hls_playlists 
    WHERE task_uuid = task_uuid_param
    ORDER BY playlist_type, resolution_config_id;
END //
DELIMITER ;

COMMIT;
-- 扩展现有转码任务表，添加HLS切片相关字段

-- 1. 扩展转码任务表
ALTER TABLE transcode_tasks 
ADD COLUMN enable_hls BOOLEAN DEFAULT FALSE COMMENT '是否启用HLS切片',
ADD COLUMN hls_resolutions JSON COMMENT 'HLS多分辨率配置 [{"resolution":"720p","bitrate":"2000k"},{"resolution":"480p","bitrate":"1000k"}]',
ADD COLUMN hls_segment_duration INT DEFAULT 10 COMMENT 'HLS切片时长(秒)',
ADD COLUMN hls_list_size INT DEFAULT 0 COMMENT 'HLS播放列表大小(0表示无限制)',
ADD COLUMN hls_format VARCHAR(20) DEFAULT 'mpegts' COMMENT 'HLS格式(mpegts/fmp4)',
ADD COLUMN hls_status VARCHAR(20) DEFAULT 'pending' COMMENT 'HLS切片状态(pending/processing/completed/failed/disabled)',
ADD COLUMN hls_progress INT DEFAULT 0 COMMENT 'HLS切片进度(0-100)',
ADD COLUMN hls_output_path VARCHAR(512) COMMENT 'HLS输出路径(master.m3u8)',
ADD COLUMN hls_error_message VARCHAR(500) COMMENT 'HLS切片错误信息',
ADD COLUMN hls_created_at TIMESTAMP NULL COMMENT 'HLS切片开始时间',
ADD COLUMN hls_completed_at TIMESTAMP NULL COMMENT 'HLS切片完成时间';

-- 2. 创建HLS切片子任务表
CREATE TABLE hls_slice_tasks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_uuid VARCHAR(36) NOT NULL COMMENT '子任务UUID',
    transcode_task_uuid VARCHAR(36) NOT NULL COMMENT '关联的转码任务UUID',
    user_uuid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    video_uuid VARCHAR(36) NOT NULL COMMENT '视频UUID',
    
    -- 分辨率配置
    resolution VARCHAR(50) NOT NULL COMMENT '分辨率(720p/480p/360p等)',
    bitrate VARCHAR(50) NOT NULL COMMENT '码率(2000k/1000k等)',
    
    -- 路径信息
    input_path VARCHAR(512) NOT NULL COMMENT '输入MP4文件路径',
    output_dir VARCHAR(512) NOT NULL COMMENT '输出目录路径',
    playlist_path VARCHAR(512) COMMENT '播放列表路径(.m3u8)',
    
    -- HLS参数
    segment_duration INT DEFAULT 10 COMMENT '切片时长(秒)',
    list_size INT DEFAULT 0 COMMENT '播放列表大小',
    format VARCHAR(20) DEFAULT 'mpegts' COMMENT 'HLS格式',
    
    -- 任务状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '任务状态(pending/processing/completed/failed)',
    progress INT DEFAULT 0 COMMENT '进度(0-100)',
    error_message VARCHAR(500) COMMENT '错误信息',
    
    -- 统计信息
    segment_count INT DEFAULT 0 COMMENT '生成的切片数量',
    total_duration DECIMAL(10,2) DEFAULT 0 COMMENT '总时长(秒)',
    file_size BIGINT DEFAULT 0 COMMENT '总文件大小(字节)',
    
    -- 时间字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL COMMENT '开始处理时间',
    completed_at TIMESTAMP NULL COMMENT '完成时间',
    
    -- 索引
    UNIQUE KEY uk_task_uuid (task_uuid),
    KEY idx_transcode_task_uuid (transcode_task_uuid),
    KEY idx_user_uuid (user_uuid),
    KEY idx_video_uuid (video_uuid),
    KEY idx_status (status),
    KEY idx_resolution (resolution),
    KEY idx_created_at (created_at),
    
    -- 外键约束
    FOREIGN KEY (transcode_task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HLS切片子任务表';

-- 3. 创建HLS切片段信息表(可选，用于详细跟踪)
CREATE TABLE hls_segments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    segment_uuid VARCHAR(36) NOT NULL COMMENT '切片UUID',
    slice_task_uuid VARCHAR(36) NOT NULL COMMENT '关联的切片任务UUID',
    transcode_task_uuid VARCHAR(36) NOT NULL COMMENT '关联的转码任务UUID',
    
    -- 切片信息
    segment_index INT NOT NULL COMMENT '切片索引',
    segment_filename VARCHAR(255) NOT NULL COMMENT '切片文件名',
    segment_path VARCHAR(512) NOT NULL COMMENT '切片文件路径',
    segment_duration DECIMAL(8,3) NOT NULL COMMENT '切片时长(秒)',
    file_size BIGINT DEFAULT 0 COMMENT '文件大小(字节)',
    
    -- 状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态(pending/processing/completed/failed)',
    error_message VARCHAR(255) COMMENT '错误信息',
    
    -- 时间字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引
    UNIQUE KEY uk_segment_uuid (segment_uuid),
    KEY idx_slice_task_uuid (slice_task_uuid),
    KEY idx_transcode_task_uuid (transcode_task_uuid),
    KEY idx_segment_index (segment_index),
    KEY idx_status (status),
    
    -- 外键约束
    FOREIGN KEY (slice_task_uuid) REFERENCES hls_slice_tasks(task_uuid) ON DELETE CASCADE,
    FOREIGN KEY (transcode_task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HLS切片段信息表';

-- 4. 创建HLS主播放列表表
CREATE TABLE hls_master_playlists (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    playlist_uuid VARCHAR(36) NOT NULL COMMENT '播放列表UUID',
    transcode_task_uuid VARCHAR(36) NOT NULL COMMENT '关联的转码任务UUID',
    user_uuid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    video_uuid VARCHAR(36) NOT NULL COMMENT '视频UUID',
    
    -- 播放列表信息
    master_m3u8_path VARCHAR(512) NOT NULL COMMENT 'master.m3u8文件路径',
    playlist_content TEXT COMMENT '播放列表内容',
    resolutions JSON COMMENT '包含的分辨率信息',
    
    -- 状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态(pending/processing/completed/failed)',
    error_message VARCHAR(255) COMMENT '错误信息',
    
    -- 时间字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引
    UNIQUE KEY uk_playlist_uuid (playlist_uuid),
    UNIQUE KEY uk_transcode_task_uuid (transcode_task_uuid),
    KEY idx_user_uuid (user_uuid),
    KEY idx_video_uuid (video_uuid),
    KEY idx_status (status),
    
    -- 外键约束
    FOREIGN KEY (transcode_task_uuid) REFERENCES transcode_tasks(task_uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HLS主播放列表表';

-- 5. 添加索引优化
CREATE INDEX idx_transcode_tasks_hls_status ON transcode_tasks(hls_status);
CREATE INDEX idx_transcode_tasks_enable_hls ON transcode_tasks(enable_hls);
CREATE INDEX idx_transcode_tasks_hls_created_at ON transcode_tasks(hls_created_at);

-- 6. 添加触发器，自动更新HLS状态时间
DELIMITER $$
CREATE TRIGGER tr_transcode_tasks_hls_status_update 
BEFORE UPDATE ON transcode_tasks
FOR EACH ROW
BEGIN
    -- 当HLS状态从非processing变为processing时，设置开始时间
    IF OLD.hls_status != 'processing' AND NEW.hls_status = 'processing' THEN
        SET NEW.hls_created_at = CURRENT_TIMESTAMP;
    END IF;
    
    -- 当HLS状态变为completed或failed时，设置完成时间
    IF OLD.hls_status NOT IN ('completed', 'failed') AND NEW.hls_status IN ('completed', 'failed') THEN
        SET NEW.hls_completed_at = CURRENT_TIMESTAMP;
    END IF;
END$$
DELIMITER ;