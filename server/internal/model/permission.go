package model

// 权限位定义（int64 位掩码）
// 与前端 permission.ts 镜像，修改时需同步
const (
	// 服务器级权限
	PermManageServer    int64 = 1 << 0  // 管理服务器设置
	PermManageRoles     int64 = 1 << 1  // 管理角色
	PermManageChannels  int64 = 1 << 2  // 管理频道
	PermKickMembers     int64 = 1 << 3  // 踢出成员
	PermBanMembers      int64 = 1 << 4  // 封禁成员
	PermInviteMembers   int64 = 1 << 5  // 邀请成员
	PermChangeNickname  int64 = 1 << 6  // 修改昵称
	PermManageNicknames int64 = 1 << 7  // 管理他人昵称

	// 文字频道权限
	PermViewChannel     int64 = 1 << 8  // 查看频道
	PermSendMessages    int64 = 1 << 9  // 发送消息
	PermManageMessages  int64 = 1 << 10 // 管理消息(删/置顶)
	PermEmbedLinks      int64 = 1 << 11 // 嵌入链接
	PermAttachFiles     int64 = 1 << 12 // 附加文件
	PermReadHistory     int64 = 1 << 13 // 读取历史消息
	PermMentionEveryone int64 = 1 << 14 // 提及 @everyone
	PermAddReactions    int64 = 1 << 15 // 添加表情反应

	// 语音频道权限
	PermConnectVoice    int64 = 1 << 16 // 连接语音
	PermSpeak           int64 = 1 << 17 // 说话
	PermMuteMembers     int64 = 1 << 18 // 静音他人
	PermDeafenMembers   int64 = 1 << 19 // 消音他人
	PermMoveMembers     int64 = 1 << 20 // 移动成员
	PermScreenShare     int64 = 1 << 21 // 屏幕共享
	PermPrioritySpeaker int64 = 1 << 22 // 优先发言

	// 管理员权限（拥有所有权限，除 owner 独占）
	PermAdministrator int64 = 1 << 30
)

// AllPermissions 所有权限位（不含 Administrator 和 owner 独占位）
const AllPermissions int64 = (1 << 30) - 1

// 默认角色权限位掩码
var (
	// DefaultOwnerPermissions owner 角色权限 = 所有权限 | Administrator
	DefaultOwnerPermissions int64 = AllPermissions | PermAdministrator
	// DefaultAdminPermissions admin 角色权限
	DefaultAdminPermissions int64 = PermManageServer |
		PermManageChannels |
		PermKickMembers |
		PermInviteMembers |
		PermChangeNickname |
		PermManageNicknames |
		PermViewChannel |
		PermSendMessages |
		PermManageMessages |
		PermReadHistory |
		PermAddReactions |
		PermConnectVoice |
		PermSpeak |
		PermMuteMembers |
		PermMoveMembers |
		PermScreenShare
	// DefaultMemberPermissions member 默认角色权限
	DefaultMemberPermissions int64 = PermViewChannel |
		PermSendMessages |
		PermReadHistory |
		PermAddReactions |
		PermConnectVoice |
		PermSpeak |
		PermScreenShare
)

// HasPermission 检查权限位是否包含指定权限
// 如果权限中包含 Administrator 位，则拥有所有权限
func HasPermission(perms int64, perm int64) bool {
	if perms&PermAdministrator != 0 {
		return true
	}
	return perms&perm != 0
}

// MergePermissions 将多个角色的权限位 OR 合并
func MergePermissions(rolePerms []int64) int64 {
	var result int64
	for _, p := range rolePerms {
		result |= p
	}
	return result
}

// IsAdministrator 检查权限中是否包含 Administrator 位
func IsAdministrator(perms int64) bool {
	return perms&PermAdministrator != 0
}
