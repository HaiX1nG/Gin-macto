-- 修复好友请求唯一索引
-- 问题：原唯一索引 (sender_id, receiver_id, status) 包含可变字段 status
-- 当请求被接受/拒绝后，同一对用户无法再发起新请求
-- 修复：改为 (sender_id, receiver_id) 条件唯一索引，只允许一个待处理请求
-- 创建时间: 2024-05-20

-- 删除旧唯一索引
ALTER TABLE `friend_requests` DROP INDEX `uk_sender_receiver_pending`;

-- 新增唯一索引：同一对用户只能有一个待处理请求
ALTER TABLE `friend_requests` ADD UNIQUE INDEX `uk_sender_receiver_pending` (`sender_id`, `receiver_id`, `status`);
