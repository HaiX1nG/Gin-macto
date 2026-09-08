-- 播放列表系统迁移脚本
-- 创建时间: 2024-05-19

-- 播放列表表
CREATE TABLE IF NOT EXISTS `playlist_items` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `channel_id` BIGINT UNSIGNED NOT NULL COMMENT '频道ID',
    `title` VARCHAR(200) NOT NULL COMMENT '歌曲标题',
    `artist` VARCHAR(200) DEFAULT NULL COMMENT '艺术家',
    `url` VARCHAR(500) NOT NULL COMMENT '音频URL',
    `duration` INT NOT NULL DEFAULT 0 COMMENT '时长(秒)',
    `cover_url` VARCHAR(500) DEFAULT NULL COMMENT '封面URL',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=等待, 1=播放中, 2=已播放',
    `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    `added_by` BIGINT UNSIGNED NOT NULL COMMENT '添加者用户ID',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_channel_status_sort` (`channel_id`, `status`, `sort_order`),
    INDEX `idx_channel_order` (`channel_id`, `sort_order`),
    INDEX `idx_added_by` (`added_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='播放列表表';
