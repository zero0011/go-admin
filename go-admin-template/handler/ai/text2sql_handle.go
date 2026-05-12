package ai

import (
	"go-admin-template/internal/response"
	logicAI "go-admin-template/logic/ai"
	"go-admin-template/svc"
	"go-admin-template/types"

	"github.com/gin-gonic/gin"
)

// Text2SQLHandle handles natural language to SQL conversion requests.
func Text2SQLHandle(c *gin.Context) {
	var req types.Text2SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := logicAI.Text2SQL(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}
