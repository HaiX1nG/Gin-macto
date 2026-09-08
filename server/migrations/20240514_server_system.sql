-- 服务器系统迁移脚本
-- 创建时间: 2024-05-14

-- 服务器表
CREATE TABLE IF NOT EXISTS `servers` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL COMMENT '服务器名称',
    `icon_url` VARCHAR(500) DEFAULT NULL COMMENT '图标URL',
    `banner_url` VARCHAR(500) DEFAULT NULL COMMENT '横幅URL',
    `description` VARCHAR(500) DEFAULT NULL COMMENT '描述',
    `owner_id` BIGINT UNSIGNED NOT NULL COMMENT '拥有者用户ID',
    `invite_code` VARCHAR(20) NOT NULL COMMENT '邀请码',
    `is_private` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否私密',
    `max_members` INT NOT NULL DEFAULT 500 COMMENT '最大成员数',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_invite_code` (`invite_code`),
    INDEX `idx_owner` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务器表';

-- 服务器成员表
CREATE TABLE IF NOT EXISTS `server_members` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `server_id` BIGINT UNSIGNED NOT NULL COMMENT '服务器ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `nickname` VARCHAR(50) DEFAULT NULL COMMENT '服务器内昵称',
    `joined_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_server_user` (`server_id`, `user_id`),
    INDEX `idx_server` (`server_id`),
    INDEX `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务器成员表';

-- 角色表
CREATE TABLE IF NOT EXISTS `roles` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `server_id` BIGINT UNSIGNED NOT NULL COMMENT '服务器ID',
    `name` VARCHAR(50) NOT NULL COMMENT '角色名称',
    `color` VARCHAR(7) NOT NULL DEFAULT '#99aab5' COMMENT '角色颜色',
    `position` INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    `permissions` BIGINT NOT NULL DEFAULT 0 COMMENT '权限位掩码',
    `is_mentionable` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否可提及',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_server_position` (`server_id`, `position`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 成员角色关联表
CREATE TABLE IF NOT EXISTS `server_member_roles` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `server_id` BIGINT UNSIGNED NOT NULL COMMENT '服务器ID',
    `member_id` BIGINT UNSIGNED NOT NULL COMMENT '成员ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_member_role` (`member_id`, `role_id`),
    INDEX `idx_server` (`server_id`),
    INDEX `idx_role` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='成员角色关联表';
