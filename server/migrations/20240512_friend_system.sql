-- 好友系统迁移脚本
-- 创建时间: 2024-05-12

-- 好友请求表
CREATE TABLE IF NOT EXISTS `friend_requests` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `sender_id` BIGINT UNSIGNED NOT NULL COMMENT '发送者ID',
    `receiver_id` BIGINT UNSIGNED NOT NULL COMMENT '接收者ID',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=待处理, 1=已接受, 2=已拒绝',
    `message` VARCHAR(200) DEFAULT '' COMMENT '请求附言',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_sender` (`sender_id`),
    INDEX `idx_receiver` (`receiver_id`),
    INDEX `idx_status` (`status`),
    UNIQUE INDEX `uk_sender_receiver_pending` (`sender_id`, `receiver_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友请求表';

-- 好友关系表
CREATE TABLE IF NOT EXISTS `friendships` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `friend_id` BIGINT UNSIGNED NOT NULL COMMENT '好友ID',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_user_friend` (`user_id`, `friend_id`),
    INDEX `idx_friend` (`friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友关系表';

-- 私聊消息表
CREATE TABLE IF NOT EXISTS `private_messages` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `sender_id` BIGINT UNSIGNED NOT NULL COMMENT '发送者ID',
    `receiver_id` BIGINT UNSIGNED NOT NULL COMMENT '接收者ID',
    `content` TEXT NOT NULL COMMENT '消息内容',
    `is_read` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已读: 0=未读, 1=已读',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_sender` (`sender_id`),
    INDEX `idx_receiver` (`receiver_id`),
    INDEX `idx_conversation` (`sender_id`, `receiver_id`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='私聊消息表';
