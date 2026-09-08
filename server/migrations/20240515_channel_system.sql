-- 频道系统迁移脚本
-- 创建时间: 2024-05-15

-- 频道表
CREATE TABLE IF NOT EXISTS `channels` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `server_id` BIGINT UNSIGNED NOT NULL COMMENT '服务器ID',
    `name` VARCHAR(100) NOT NULL COMMENT '频道名称',
    `type` TINYINT NOT NULL COMMENT '频道类型: 1=文字, 2=语音, 3=分类',
    `topic` VARCHAR(500) DEFAULT NULL COMMENT '频道主题',
    `parent_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '父分类ID',
    `position` INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    `bitrate` INT NOT NULL DEFAULT 64000 COMMENT '语音频道比特率',
    `user_limit` INT NOT NULL DEFAULT 0 COMMENT '语音频道人数上限',
    `slow_mode` INT NOT NULL DEFAULT 0 COMMENT '慢速模式(秒)',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_server_parent_position` (`server_id`, `parent_id`, `position`),
    INDEX `idx_parent` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='频道表';
