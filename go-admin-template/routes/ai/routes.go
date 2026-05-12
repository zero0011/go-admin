package ai

import (
	handlerAI "go-admin-template/handler/ai"
	"go-admin-template/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAIRoute(e *gin.Engine) {
	g := e.Group("/ai")
	g.Use(middleware.JwtMiddleware, middleware.AuthMiddleware)
	g.POST("/text2sql", handlerAI.Text2SQLHandle)
}
