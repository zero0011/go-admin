package file

import (
	"go-admin-template/handler/file"
	"go-admin-template/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterFileRoute(e *gin.Engine) {
	g := e.Group("")
	g.Use(middleware.JwtMiddleware)

	// 文件夹管理
	g.GET("/file/folder/list", file.GetFolderListHandle)
	g.POST("/file/folder", file.CreateFolderHandle)
	g.PUT("/file/folder/:id", file.RenameFolderHandle)
	g.DELETE("/file/folder/:id", file.DeleteFolderHandle)

	// 文件管理
	g.GET("/file/list", file.GetFileListHandle)
	g.GET("/file/search", file.SearchFilesHandle)
	g.POST("/file/upload", file.UploadFileHandle)
	g.DELETE("/file/:id", file.DeleteFileHandle)
	g.POST("/file/batch-delete", file.BatchDeleteFilesHandle)

	// 分片上传
	g.POST("/file/chunk/init", file.InitChunkUploadHandle)
	g.POST("/file/chunk/upload", file.UploadChunkHandle)
	g.POST("/file/chunk/merge", file.MergeChunkHandle)
	g.GET("/file/chunk/status", file.GetChunkStatusHandle)
}
