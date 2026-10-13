package message

import (
	"go-admin-template/handler/message"
	"go-admin-template/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterMessageRoute 站内信与实时公告路由
// 用户侧仅挂 JwtMiddleware；运营侧挂 Jwt + Casbin（沿用现有角色，role:1 旁路放行，不新增权限点）
func RegisterMessageRoute(e *gin.Engine) {
	user := e.Group("")
	user.Use(middleware.JwtMiddleware)

	// 用户侧
	user.GET("/message/list", message.ListMessagesHandle)
	user.GET("/message/unread-count", message.UnreadCountHandle)
	user.POST("/message/read", message.MarkReadHandle)
	user.POST("/message/read-all", message.MarkAllReadHandle)
	user.GET("/message/stream", message.StreamHandle)

	ops := e.Group("")
	ops.Use(middleware.JwtMiddleware, middleware.AuthMiddleware)

	// 运营侧
	ops.POST("/message/announcement", message.CreateAnnouncementHandle)
	ops.GET("/message/announcement/list", message.AnnouncementListHandle)
}
