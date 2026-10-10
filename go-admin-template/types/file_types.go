package types

import "mime/multipart"

// ========== 文件夹相关类型 ==========

// FileFolderCreateRequest 创建文件夹请求
type FileFolderCreateRequest struct {
	FolderName string `json:"folder_name" binding:"required,max=255" label:"文件夹名称"`
	ParentID   *int   `json:"parent_id" label:"父文件夹ID"`
}

// FileFolderRenameRequest 重命名文件夹请求
type FileFolderRenameRequest struct {
	FolderName string `json:"folder_name" binding:"required,max=255" label:"文件夹名称"`
}

// FileFolderListRequest 获取文件夹列表请求
type FileFolderListRequest struct {
	ParentID *int   `form:"parent_id" label:"父文件夹ID"`
	Keyword  string `form:"keyword" label:"搜索关键词"`
}

// FileFolderItem 文件夹项
type FileFolderItem struct {
	ID         int       `json:"id"`
	FolderName string    `json:"folder_name"`
	ParentID   *int      `json:"parent_id"`
	CreatedAt  LocalTime `json:"created_at"`
}

// FileFolderListResponse 获取文件夹列表响应
type FileFolderListResponse struct {
	Data []FileFolderItem `json:"data"`
}

// ========== 文件相关类型 ==========

// FileType 文件类型
type FileType string

const (
	FileTypeImage    FileType = "image"    // 图片
	FileTypeVideo    FileType = "video"    // 视频
	FileTypeAudio    FileType = "audio"    // 音频
	FileTypeDocument FileType = "document" // 文档
	FileTypeOther    FileType = "other"    // 其他
)

// SortBy 排序方式
type SortBy string

const (
	SortByNameAsc  SortBy = "name_asc"
	SortByNameDesc SortBy = "name_desc"
	SortByTimeAsc  SortBy = "time_asc"
	SortByTimeDesc SortBy = "time_desc"
	SortBySizeAsc  SortBy = "size_asc"
	SortBySizeDesc SortBy = "size_desc"
)

// FileListRequest 获取文件列表请求
type FileListRequest struct {
	ParentFolderID *int    `form:"parent_folder_id" label:"父文件夹ID"`
	FileType       *string `form:"file_type" label:"文件类型"`
	SortBy         string  `form:"sort_by" label:"排序方式"`
	Keyword        string  `form:"keyword" label:"搜索关键词"`
	Page           int     `form:"page" label:"页码"`
	PageSize       int     `form:"page_size" label:"每页条数"`
}

// FileSearchRequest 全局搜索请求
type FileSearchRequest struct {
	Keyword  string  `form:"keyword" binding:"required,max=255" label:"搜索关键词"`
	FileType *string `form:"file_type" label:"文件类型"`
	SortBy   string  `form:"sort_by" label:"排序方式"`
	Page     int     `form:"page" label:"页码"`
	PageSize int     `form:"page_size" label:"每页条数"`
}

// FilePathItem 路径节点
type FilePathItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// FileItem 文件项
type FileItem struct {
	AssetID        int       `json:"asset_id"`
	FileName       string    `json:"file_name"`
	FileURL        string    `json:"file_url"`
	FileSize       int64     `json:"file_size"`
	FileExt        string    `json:"file_ext"`
	ParentFolderID *int      `json:"parent_folder_id"`
	CreatedAt      LocalTime `json:"created_at"`
}

// FileListItem 文件/文件夹混合列表项
type FileListItem struct {
	Type           string         `json:"type"` // "folder" 或 "asset"
	FolderID       *int           `json:"folder_id,omitempty"`
	FolderName     string         `json:"folder_name,omitempty"`
	ParentFolderID *int           `json:"parent_folder_id,omitempty"`
	AssetID        *int           `json:"asset_id,omitempty"`
	FileName       string         `json:"file_name,omitempty"`
	FileURL        string         `json:"file_url,omitempty"`
	FileSize       *int64         `json:"file_size,omitempty"`
	FileExt        string         `json:"file_ext,omitempty"`
	CreatedAt      *string        `json:"created_at,omitempty"`
	Path           []FilePathItem `json:"path,omitempty"` // 所在目录路径（不含自身），仅搜索结果返回
}

// FileListResponse 获取文件列表响应
type FileListResponse struct {
	Total int            `json:"total"`
	List  []FileListItem `json:"list"`
}

// ========== 单文件上传相关类型 ==========

// FileUploadRequest 单文件上传请求
type FileUploadRequest struct {
	File           *multipart.FileHeader `form:"file" file:"-" binding:"required" label:"文件"`
	ParentFolderID *int                  `form:"parent_folder_id" label:"父文件夹ID"`
}

// FileUploadResponse 单文件上传响应
type FileUploadResponse struct {
	AssetID  int    `json:"asset_id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
}

// ========== 分片上传相关类型 ==========

// ChunkInitRequest 分片上传初始化请求
type ChunkInitRequest struct {
	FileName       string `json:"file_name" binding:"required" label:"文件名"`
	FileSize       int64  `json:"file_size" binding:"required,min=1" label:"文件大小"`
	ChunkSize      int    `json:"chunk_size" binding:"required,min=1" label:"分片大小"`
	FileHash       string `json:"file_hash" binding:"required" label:"文件哈希"`
	ParentFolderID *int   `json:"parent_folder_id" label:"父文件夹ID"`
}

// ChunkInitResponse 分片上传初始化响应
type ChunkInitResponse struct {
	UploadID   string `json:"upload_id"`
	ChunkCount int    `json:"chunk_count"`
	ChunkSize  int    `json:"chunk_size"`
}

// ChunkUploadRequest 分片上传请求
type ChunkUploadRequest struct {
	UploadID   string                `form:"upload_id" binding:"required" label:"上传会话ID"`
	ChunkIndex int                   `form:"chunk_index" binding:"required,min=0" label:"分片索引"`
	ChunkHash  string                `form:"chunk_hash" binding:"required" label:"分片哈希"`
	File       *multipart.FileHeader `form:"file" file:"-" binding:"required" label:"分片文件"`
}

// ChunkUploadResponse 分片上传响应
type ChunkUploadResponse struct {
	ChunkIndex int  `json:"chunk_index"`
	Uploaded   bool `json:"uploaded"`
}

// ChunkMergeRequest 分片合并请求
type ChunkMergeRequest struct {
	UploadID       string `json:"upload_id" binding:"required" label:"上传会话ID"`
	ParentFolderID *int   `json:"parent_folder_id" label:"父文件夹ID"`
}

// ChunkMergeResponse 分片合并响应
type ChunkMergeResponse struct {
	AssetID  int    `json:"asset_id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
}

// ChunkStatusRequest 查询分片状态请求
type ChunkStatusRequest struct {
	UploadID string `form:"upload_id" binding:"required" label:"上传会话ID"`
}

// ChunkStatusResponse 查询分片状态响应
type ChunkStatusResponse struct {
	UploadID       string `json:"upload_id"`
	UploadedChunks []int  `json:"uploaded_chunks"`
	TotalChunks    int    `json:"total_chunks"`
}

// ========== 文件操作相关类型 ==========

// FileDeleteRequest 删除文件请求
type FileDeleteRequest struct {
	ID int `uri:"id" binding:"required" label:"文件ID"`
}

// FileBatchDeleteRequest 批量删除文件请求
type FileBatchDeleteRequest struct {
	IDs []int `json:"ids" binding:"required,min=1" label:"文件ID列表"`
}
