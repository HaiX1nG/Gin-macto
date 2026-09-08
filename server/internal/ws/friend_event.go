package ws

import (
	"context"
	"time"

	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/repository"
)

// NotifyFriendOnline 通知该用户的所有好友上线
// 通过 friendRepo 查询该用户的所有好友，然后使用 hub.SendToUser 逐个推送上线通知
// 参数：
//   - hub: WebSocket Hub 实例
//   - friendRepo: 好友仓储实例
//   - userID: 上线用户ID
func NotifyFriendOnline(hub *Hub, friendRepo *repository.FriendRepository, userID uint64) {
	if hub == nil || friendRepo == nil || userID == 0 {
		return
	}

	// 查询该用户的好友列表（使用较大的分页限制以获取所有好友）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	friendships, _, err := friendRepo.FindFriendsByUserID(ctx, userID, 1, 1000)
	if err != nil {
		return
	}

	// 向所有好友推送上线通知
	for _, f := range friendships {
		hub.SendToUser(f.FriendID, EventFriendOnline, map[string]any{
			"userId": userID,
		})
	}
}

// NotifyFriendOffline 通知该用户的所有好友下线
// 通过 friendRepo 查询该用户的所有好友，然后使用 hub.SendToUser 逐个推送下线通知
// 参数：
//   - hub: WebSocket Hub 实例
//   - friendRepo: 好友仓储实例
//   - userID: 下线用户ID
func NotifyFriendOffline(hub *Hub, friendRepo *repository.FriendRepository, userID uint64) {
	if hub == nil || friendRepo == nil || userID == 0 {
		return
	}

	// 查询该用户的好友列表
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	friendships, _, err := friendRepo.FindFriendsByUserID(ctx, userID, 1, 1000)
	if err != nil {
		return
	}

	// 向所有好友推送下线通知
	for _, f := range friendships {
		hub.SendToUser(f.FriendID, EventFriendOffline, map[string]any{
			"userId": userID,
		})
	}
}

// NotifyFriendRequest 推送好友请求给接收者
// 当有用户发送好友请求时，通过 WebSocket 实时推送通知给接收者
// 参数：
//   - hub: WebSocket Hub 实例
//   - friendRepo: 好友仓储实例（预留，用于未来扩展）
//   - receiverID: 接收者用户ID
//   - request: 好友请求响应数据
func NotifyFriendRequest(hub *Hub, friendRepo *repository.FriendRepository, receiverID uint64, request *dto.FriendRequestResponse) {
	if hub == nil || receiverID == 0 || request == nil {
		return
	}

	hub.SendToUser(receiverID, EventFriendRequest, map[string]any{
		"request": request,
	})
}

// NotifyFriendRequestResult 通知请求处理结果
// 当好友请求被接受或拒绝时，推送通知给请求发起者
// 参数：
//   - hub: WebSocket Hub 实例
//   - friendRepo: 好友仓储实例（预留，用于未来扩展）
//   - userID: 请求发起者用户ID
//   - accepted: 是否被接受
//   - friendID: 处理请求的用户ID
//   - friendName: 处理请求的用户名称
func NotifyFriendRequestResult(hub *Hub, friendRepo *repository.FriendRepository, userID uint64, accepted bool, friendID uint64, friendName string) {
	if hub == nil || userID == 0 {
		return
	}

	hub.SendToUser(userID, EventFriendRequestResult, map[string]any{
		"accepted":   accepted,
		"friendId":   friendID,
		"friendName": friendName,
	})
}

// NotifyFriendAdded 通知双方好友关系建立
// 当好友请求被接受后，向双方推送好友关系建立通知
// 参数：
//   - hub: WebSocket Hub 实例
//   - friendRepo: 好友仓储实例
//   - userID: 当前用户ID
//   - friendID: 好友用户ID
//   - friendName: 好友用户名称
func NotifyFriendAdded(hub *Hub, friendRepo *repository.FriendRepository, userID, friendID uint64, friendName string) {
	if hub == nil || userID == 0 || friendID == 0 {
		return
	}

	hub.SendToUser(userID, EventFriendAdded, map[string]any{
		"friendId":   friendID,
		"friendName": friendName,
	})
}

// NotifyFriendRemoved 通知双方好友关系解除
// 当删除好友后，向双方推送好友关系解除通知
// 参数：
//   - hub: WebSocket Hub 实例
//   - friendRepo: 好友仓储实例
//   - userID: 当前用户ID
//   - friendID: 被删除的好友用户ID
func NotifyFriendRemoved(hub *Hub, friendRepo *repository.FriendRepository, userID, friendID uint64) {
	if hub == nil || userID == 0 || friendID == 0 {
		return
	}

	// 向当前用户推送好友关系解除通知
	hub.SendToUser(userID, EventFriendRemoved, map[string]any{
		"friendId": friendID,
	})

	// 向被删除的好友推送好友关系解除通知
	hub.SendToUser(friendID, EventFriendRemoved, map[string]any{
		"friendId": userID,
	})
}
