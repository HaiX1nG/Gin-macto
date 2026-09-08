-- 消息系统迁移脚本
-- 创建时间: 2024-05-16

-- 频道消息表
CREATE TABLE IF NOT EXISTS `channel_messages` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `channel_id` BIGINT UNSIGNED NOT NULL COMMENT '频道ID',
    `sender_user_id` BIGINT UNSIGNED NOT NULL COMMENT '发送者用户ID',
    `type` TINYINT NOT NULL DEFAULT 1 COMMENT '消息类型: 1=文本, 2=图片, 3=系统',
    `content` TEXT NOT NULL COMMENT '消息内容',
    `reply_to_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '回复消息ID',
    `edited_at` DATETIME DEFAULT NULL COMMENT '编辑时间',
    `is_pinned` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否置顶',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_channel_created` (`channel_id`, `created_at`),
    INDEX `idx_sender` (`sender_user_id`),
    INDEX `idx_reply_to` (`reply_to_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='频道消息表';

-- 消息附件表
CREATE TABLE IF NOT EXISTS `message_attachments` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `message_id` BIGINT UNSIGNED NOT NULL COMMENT '消息ID',
    `filename` VARCHAR(255) NOT NULL COMMENT '文件名',
    `url` VARCHAR(500) NOT NULL COMMENT '文件URL',
    `file_size` INT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
    `mime_type` VARCHAR(100) DEFAULT NULL COMMENT 'MIME类型',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_message` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息附件表';

-- 消息表情反应表
CREATE TABLE IF NOT EXISTS `message_reactions` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `message_id` BIGINT UNSIGNED NOT NULL COMMENT '消息ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `emoji` VARCHAR(32) NOT NULL COMMENT 'emoji字符',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_message_user_emoji` (`message_id`, `user_id`, `emoji`),
    INDEX `idx_message` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表情反应表';
