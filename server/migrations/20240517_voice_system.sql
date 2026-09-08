-- 语音与屏幕共享系统迁移脚本
-- 创建时间: 2024-05-17

-- 语音参与者表
CREATE TABLE IF NOT EXISTS `voice_participants` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `channel_id` BIGINT UNSIGNED NOT NULL COMMENT '频道ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `is_muted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否静音',
    `is_deafened` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否消音',
    `is_speaking` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否正在说话',
    `volume` INT NOT NULL DEFAULT 100 COMMENT '音量',
    `joined_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_channel_user` (`channel_id`, `user_id`),
    INDEX `idx_channel` (`channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='语音参与者表';

-- 语音会话历史表
CREATE TABLE IF NOT EXISTS `voice_sessions` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `channel_id` BIGINT UNSIGNED NOT NULL COMMENT '频道ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `joined_at` DATETIME NOT NULL COMMENT '加入时间',
    `left_at` DATETIME DEFAULT NULL COMMENT '离开时间',
    PRIMARY KEY (`id`),
    INDEX `idx_channel_user` (`channel_id`, `user_id`),
    INDEX `idx_joined_at` (`joined_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='语音会话历史表';

-- 屏幕共享会话表
CREATE TABLE IF NOT EXISTS `screen_share_sessions` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `channel_id` BIGINT UNSIGNED NOT NULL COMMENT '频道ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `started_at` DATETIME NOT NULL COMMENT '开始时间',
    `ended_at` DATETIME DEFAULT NULL COMMENT '结束时间',
    PRIMARY KEY (`id`),
    INDEX `idx_channel` (`channel_id`),
    INDEX `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='屏幕共享会话表';
