package file

import (
	"go-admin-template/internal/response"
	"go-admin-template/logic/file"
	"go-admin-template/svc"
	"go-admin-template/types"

	"github.com/gin-gonic/gin"
)

// ========== 文件夹管理 Handler ==========

// GetFolderListHandle 获取文件夹列表
func GetFolderListHandle(c *gin.Context) {
	var req types.FileFolderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.GetFolderList(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// CreateFolderHandle 创建文件夹
func CreateFolderHandle(c *gin.Context) {
	var req types.FileFolderCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.CreateFolder(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// RenameFolderHandle 重命名文件夹
func RenameFolderHandle(c *gin.Context) {
	var id types.PathID
	if err := c.ShouldBindUri(&id); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}

	var req types.FileFolderRenameRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}

	err := file.RenameFolder(svc.NewServiceContext(c), id.ID, &req)
	response.HandleResponse(c, nil, err)
}

// DeleteFolderHandle 删除文件夹
func DeleteFolderHandle(c *gin.Context) {
	var id types.PathID
	if err := c.ShouldBindUri(&id); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}

	err := file.DeleteFolder(svc.NewServiceContext(c), id.ID)
	response.HandleResponse(c, nil, err)
}

// ========== 文件管理 Handler ==========

// GetFileListHandle 获取文件列表
func GetFileListHandle(c *gin.Context) {
	var req types.FileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.GetFileList(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// SearchFilesHandle 全局搜索文件/文件夹
func SearchFilesHandle(c *gin.Context) {
	var req types.FileSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.SearchFiles(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// UploadFileHandle 单文件上传
func UploadFileHandle(c *gin.Context) {
	var req types.FileUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.UploadFile(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// DeleteFileHandle 删除文件
func DeleteFileHandle(c *gin.Context) {
	var id types.PathID
	if err := c.ShouldBindUri(&id); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}

	err := file.DeleteFile(svc.NewServiceContext(c), id.ID)
	response.HandleResponse(c, nil, err)
}

// BatchDeleteFilesHandle 批量删除文件
func BatchDeleteFilesHandle(c *gin.Context) {
	var req types.FileBatchDeleteRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}

	err := file.BatchDeleteFiles(svc.NewServiceContext(c), req.IDs)
	response.HandleResponse(c, nil, err)
}

// ========== 分片上传 Handler ==========

// InitChunkUploadHandle 初始化分片上传
func InitChunkUploadHandle(c *gin.Context) {
	var req types.ChunkInitRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.InitChunkUpload(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// UploadChunkHandle 上传分片
func UploadChunkHandle(c *gin.Context) {
	var req types.ChunkUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.UploadChunk(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// MergeChunkHandle 合并分片
func MergeChunkHandle(c *gin.Context) {
	var req types.ChunkMergeRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.MergeChunk(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}

// GetChunkStatusHandle 获取分片上传状态
func GetChunkStatusHandle(c *gin.Context) {
	var req types.ChunkStatusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.GetChunkStatus(svc.NewServiceContext(c), req.UploadID)
	response.HandleResponse(c, resp, err)
}
