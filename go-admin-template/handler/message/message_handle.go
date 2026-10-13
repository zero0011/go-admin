package message

import (
	"go-admin-template/internal/response"
	"go-admin-template/logic/message"
	"go-admin-template/pkg/jwt"
	"go-admin-template/pkg/sse"
	"go-admin-template/svc"
	"go-admin-template/types"

	"github.com/gin-gonic/gin"
)

// getUserID 从上下文获取当前用户ID
func getUserID(c *gin.Context) int {
	user, _ := c.Get(jwt.UserInfo)
	if claims, ok := user.(*jwt.Claims); ok {
		return claims.Id
	}
	return 0
}

// ========== 用户侧 ==========

// ListMessagesHandle 消息列表
func ListMessagesHandle(c *gin.Context) {
	var req types.MessageListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := message.ListMessages(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// UnreadCountHandle 未读数
func UnreadCountHandle(c *gin.Context) {
	resp, err := message.UnreadCount(svc.NewServiceContext(c))
	response.HandleResponse(c, resp, err)
}

// MarkReadHandle 单条已读
func MarkReadHandle(c *gin.Context) {
	var req types.MessageReadRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := message.MarkRead(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// MarkAllReadHandle 一键已读
func MarkAllReadHandle(c *gin.Context) {
	resp, err := message.MarkAllRead(svc.NewServiceContext(c))
	response.HandleResponse(c, resp, err)
}

// StreamHandle SSE 推送通道
// 鉴权：JwtMiddleware 支持 ?Authorization=<jwt>（EventSource 无法自定义 Header）
func StreamHandle(c *gin.Context) {
	userID := getUserID(c)
	sse.Stream(c, userID, func(_ *sse.Client) {
		// 连接建立（含重连）后立即推送当前未读数，做断线补偿
		if resp, err := message.UnreadCountForUser(userID); err == nil {
			sse.DefaultHub.PushToUser(userID, sse.EventUnread, resp)
		}
	})
}

// ========== 运营侧 ==========

// CreateAnnouncementHandle 发布公告
func CreateAnnouncementHandle(c *gin.Context) {
	var req types.MessageAnnouncementCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	err := message.CreateAnnouncement(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, nil, err)
}

// AnnouncementListHandle 已发布公告列表
func AnnouncementListHandle(c *gin.Context) {
	var req types.MessageAnnouncementListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := message.AnnouncementList(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}
