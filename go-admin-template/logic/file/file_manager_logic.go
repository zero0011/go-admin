package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-admin-template/config"
	"go-admin-template/model"
	"go-admin-template/pkg/jwt"
	"go-admin-template/svc"
	"go-admin-template/types"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// GetFolderList 获取文件夹列表
func GetFolderList(ctx *svc.ServiceContext, req *types.FileFolderListRequest) ([]types.FileFolderItem, error) {
	db := model.DB()

	query := db.Model(&model.FileFolder{})

	// 如果传入了父文件夹ID，则只查询该文件夹下的子文件夹
	if req.ParentID != nil {
		query = query.Where("parent_id = ?", *req.ParentID)
	} else {
		// 默认查询根目录下的文件夹
		query = query.Where("parent_id IS NULL")
	}

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where("folder_name LIKE ?", "%"+req.Keyword+"%")
	}

	var folders []model.FileFolder
	err := query.Order("create_time DESC").Find(&folders).Error
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询文件夹列表失败")
	}

	// 转换为响应类型
	result := make([]types.FileFolderItem, 0, len(folders))
	for _, f := range folders {
		result = append(result, types.FileFolderItem{
			ID:         f.ID,
			FolderName: f.FolderName,
			ParentID:   f.ParentID,
			CreatedAt:  types.LocalTime(f.CreateTime),
		})
	}

	return result, nil
}

// CreateFolder 创建文件夹
func CreateFolder(ctx *svc.ServiceContext, req *types.FileFolderCreateRequest) (int, error) {
	db := model.DB()

	// 获取当前用户ID
	userID := getUserID(ctx)

	folder := model.FileFolder{
		FolderName: req.FolderName,
		ParentID:   req.ParentID,
		CreatedBy:  userID,
	}

	err := db.Create(&folder).Error
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return 0, errors.New("创建文件夹失败")
	}

	return folder.ID, nil
}

// RenameFolder 重命名文件夹
func RenameFolder(ctx *svc.ServiceContext, id int, req *types.FileFolderRenameRequest) error {
	db := model.DB()

	result := db.Model(&model.FileFolder{}).Where("id = ?", id).Update("folder_name", req.FolderName)
	if result.Error != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(result.Error))
		return errors.New("重命名文件夹失败")
	}

	if result.RowsAffected == 0 {
		return errors.New("文件夹不存在")
	}

	return nil
}

// DeleteFolder 删除文件夹（级联删除子文件夹和文件）
func DeleteFolder(ctx *svc.ServiceContext, id int) error {
	db := model.DB()

	// 开启事务
	tx := db.Begin()
	if tx.Error != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(tx.Error))
		return errors.New("系统错误")
	}

	// 递归删除文件夹及其内容
	err := deleteFolderRecursive(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return errors.New("删除文件夹失败")
	}

	return nil
}

// deleteFolderRecursive 递归删除文件夹
func deleteFolderRecursive(tx *gorm.DB, folderID int) error {
	// 1. 先删除该文件夹下的所有文件
	if err := tx.Where("folder_id = ?", folderID).Delete(&model.SysFile{}).Error; err != nil {
		return errors.New("删除文件夹内的文件失败")
	}

	// 2. 查找所有子文件夹
	var subFolders []model.FileFolder
	if err := tx.Where("parent_id = ?", folderID).Find(&subFolders).Error; err != nil {
		return errors.New("查询子文件夹失败")
	}

	// 3. 递归删除子文件夹
	for _, subFolder := range subFolders {
		if err := deleteFolderRecursive(tx, subFolder.ID); err != nil {
			return err
		}
	}

	// 4. 删除当前文件夹
	if err := tx.Delete(&model.FileFolder{}, folderID).Error; err != nil {
		return errors.New("删除文件夹失败")
	}

	return nil
}

// ========== 文件管理 Logic ==========

// GetFileList 获取文件列表（与文件夹混合）
func GetFileList(ctx *svc.ServiceContext, req *types.FileListRequest) (*types.FileListResponse, error) {
	db := model.DB()

	// 默认分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	// 构建查询
	folderQuery := db.Model(&model.FileFolder{})
	fileQuery := db.Model(&model.SysFile{})

	// 文件夹查询条件
	if req.ParentFolderID != nil {
		folderQuery = folderQuery.Where("parent_id = ?", *req.ParentFolderID)
		fileQuery = fileQuery.Where("folder_id = ?", *req.ParentFolderID)
	} else {
		folderQuery = folderQuery.Where("parent_id IS NULL")
		fileQuery = fileQuery.Where("folder_id IS NULL")
	}

	// 关键词搜索（文件夹和文件名）
	if req.Keyword != "" {
		folderQuery = folderQuery.Where("folder_name LIKE ?", "%"+req.Keyword+"%")
		fileQuery = fileQuery.Where("file_name LIKE ?", "%"+req.Keyword+"%")
	}

	// 文件类型筛选
	if req.FileType != nil && *req.FileType != "" {
		fileQuery = fileQuery.Where("file_ext = ?", getExtByFileType(*req.FileType))
	}

	// 统计数量
	var folderCount, fileCount int64
	folderQuery.Count(&folderCount)
	fileQuery.Count(&fileCount)

	// 构建排序
	folderOrder := "create_time DESC"
	fileOrder := "create_time DESC"
	switch types.SortBy(req.SortBy) {
	case types.SortByNameAsc:
		folderOrder = "folder_name ASC"
		fileOrder = "file_name ASC"
	case types.SortByNameDesc:
		folderOrder = "folder_name DESC"
		fileOrder = "file_name DESC"
	case types.SortByTimeAsc:
		folderOrder = "create_time ASC"
		fileOrder = "create_time ASC"
	case types.SortByTimeDesc:
		folderOrder = "create_time DESC"
		fileOrder = "create_time DESC"
	case types.SortBySizeAsc:
		fileOrder = "file_size ASC"
	case types.SortBySizeDesc:
		fileOrder = "file_size DESC"
	}

	// 查询文件夹
	var folders []model.FileFolder
	folderQuery.Order(folderOrder).Find(&folders)

	// 查询文件（分页）
	var files []model.SysFile
	offset := (req.Page - 1) * req.PageSize
	fileQuery.Order(fileOrder).Offset(offset).Limit(req.PageSize).Find(&files)

	// 组装混合列表
	result := make([]types.FileListItem, 0, len(folders)+len(files))

	// 添加文件夹
	for _, f := range folders {
		result = append(result, types.FileListItem{
			Type:           "folder",
			FolderID:       lo.ToPtr(f.ID),
			FolderName:     f.FolderName,
			ParentFolderID: f.ParentID,
		})
	}

	// 添加文件
	for _, f := range files {
		result = append(result, types.FileListItem{
			Type:           "asset",
			AssetID:        lo.ToPtr(f.ID),
			FileName:       f.FileName,
			FileURL:        f.FilePath,
			FileSize:       lo.ToPtr(f.FileSize),
			FileExt:        f.FileExt,
			ParentFolderID: f.FolderID,
			CreatedAt:      lo.ToPtr(f.CreateTime.Format("2006-01-02 15:04:05")),
		})
	}

	return &types.FileListResponse{
		Total: int(folderCount + fileCount),
		List:  result,
	}, nil
}

// getExtByFileType 根据文件类型获取扩展名
func getExtByFileType(fileType string) string {
	extMap := map[string]string{
		"image":    ".image",
		"video":    ".video",
		"audio":    ".audio",
		"document": ".document",
	}
	if ext, ok := extMap[fileType]; ok {
		return ext
	}
	return ""
}

// UploadFile 单文件上传
func UploadFile(ctx *svc.ServiceContext, req *types.FileUploadRequest) (*types.FileUploadResponse, error) {
	db := model.DB()

	// 获取文件
	file, err := req.File.Open()
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("无法读取文件")
	}
	defer file.Close()

	// 生成唯一文件名
	originalName := req.File.Filename
	ext := strings.ToLower(filepath.Ext(originalName))
	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), generateRandomString(8), ext)

	// 获取当前用户ID
	userID := getUserID(ctx)

	// 创建存储目录
	now := time.Now()
	storeDir := filepath.Join("uploads", "files", fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()))
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("创建存储目录失败")
	}

	// 保存文件
	storePath := filepath.Join(storeDir, uniqueName)
	dst, err := os.Create(storePath)
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("保存文件失败")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("保存文件失败")
	}

	// 保存文件记录到数据库
	sysFile := model.SysFile{
		FileName:     originalName,
		FilePath:     storePath,
		FileSize:     req.File.Size,
		FileExt:      ext,
		FolderID:     req.ParentFolderID,
		UploadUserID: userID,
	}

	if err := db.Create(&sysFile).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		// 删除已保存的文件
		os.Remove(storePath)
		return nil, errors.New("保存文件记录失败")
	}

	// 返回完整URL
	fileURL := config.ServerConf.BaseUrl + "/" + storePath

	return &types.FileUploadResponse{
		AssetID:  sysFile.ID,
		FileName: sysFile.FileName,
		FileURL:  fileURL,
		FileSize: sysFile.FileSize,
	}, nil
}

// DeleteFile 删除文件
func DeleteFile(ctx *svc.ServiceContext, id int) error {
	db := model.DB()

	var file model.SysFile
	if err := db.First(&file, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件不存在")
		}
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return errors.New("删除文件失败")
	}

	// 删除物理文件
	if err := os.Remove(file.FilePath); err != nil {
		ctx.Log.Warnf("删除物理文件失败: %s", err.Error())
		// 继续删除数据库记录
	}

	// 删除数据库记录
	if err := db.Delete(&file).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return errors.New("删除文件失败")
	}

	return nil
}

// BatchDeleteFiles 批量删除文件
func BatchDeleteFiles(ctx *svc.ServiceContext, ids []int) error {
	db := model.DB()

	var files []model.SysFile
	if err := db.Where("id IN ?", ids).Find(&files).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return errors.New("删除文件失败")
	}

	// 删除物理文件
	for _, file := range files {
		if err := os.Remove(file.FilePath); err != nil {
			ctx.Log.Warnf("删除物理文件失败: %s", err.Error())
		}
	}

	// 批量删除数据库记录
	if err := db.Where("id IN ?", ids).Delete(&model.SysFile{}).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return errors.New("删除文件失败")
	}

	return nil
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}
	return string(result)
}

// getUserID 从上下文中获取当前用户ID
func getUserID(ctx *svc.ServiceContext) int {
	user, _ := ctx.GinContext.Get(jwt.UserInfo)
	if claims, ok := user.(*jwt.Claims); ok {
		return claims.Id
	}
	return 0
}

// ========== 分片上传 Logic ==========

// ChunkUploadSession 分片上传会话
type ChunkUploadSession struct {
	UploadID       string
	FileName       string
	FileSize       int64
	ChunkSize      int
	FileHash       string
	ParentFolderID *int
	ChunkCount     int
	UploadUserID   int
}

// sessionStore 会话存储（内存中，生产环境应使用Redis）
var sessionStore = make(map[string]*ChunkUploadSession)

// InitChunkUpload 初始化分片上传
func InitChunkUpload(ctx *svc.ServiceContext, req *types.ChunkInitRequest) (*types.ChunkInitResponse, error) {
	// 计算分片数量
	chunkCount := int(req.FileSize) / req.ChunkSize
	if int(req.FileSize)%req.ChunkSize > 0 {
		chunkCount++
	}

	// 生成上传会话ID
	uploadID := uuid.New().String()

	// 保存会话信息
	userID := getUserID(ctx)
	sessionStore[uploadID] = &ChunkUploadSession{
		UploadID:       uploadID,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		ChunkSize:      req.ChunkSize,
		FileHash:       req.FileHash,
		ParentFolderID: req.ParentFolderID,
		ChunkCount:     chunkCount,
		UploadUserID:   userID,
	}

	return &types.ChunkInitResponse{
		UploadID:   uploadID,
		ChunkCount: chunkCount,
		ChunkSize:  req.ChunkSize,
	}, nil
}

// UploadChunk 上传分片
func UploadChunk(ctx *svc.ServiceContext, req *types.ChunkUploadRequest) (*types.ChunkUploadResponse, error) {
	// 获取会话信息
	if _, ok := sessionStore[req.UploadID]; !ok {
		return nil, errors.New("上传会话不存在或已过期")
	}

	// 获取分片文件
	file, err := req.File.Open()
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("无法读取分片文件")
	}
	defer file.Close()

	// 创建分片存储目录
	now := time.Now()
	chunkDir := filepath.Join("uploads", "chunks", fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()), req.UploadID)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("创建分片存储目录失败")
	}

	// 保存分片文件
	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("part_%d", req.ChunkIndex))
	dst, err := os.Create(chunkPath)
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("保存分片文件失败")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("保存分片文件失败")
	}

	// 保存分片记录到数据库
	db := model.DB()
	chunkRecord := model.SysFileChunk{
		UploadID:   req.UploadID,
		ChunkIndex: req.ChunkIndex,
		ChunkHash:  req.ChunkHash,
		ChunkPath:  chunkPath,
	}

	if err := db.Create(&chunkRecord).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("保存分片记录失败")
	}

	return &types.ChunkUploadResponse{
		ChunkIndex: req.ChunkIndex,
		Uploaded:   true,
	}, nil
}

// MergeChunk 合并分片
func MergeChunk(ctx *svc.ServiceContext, req *types.ChunkMergeRequest) (*types.ChunkMergeResponse, error) {
	// 获取会话信息
	session, ok := sessionStore[req.UploadID]
	if !ok {
		return nil, errors.New("上传会话不存在或已过期")
	}

	db := model.DB()

	// 检查所有分片是否已上传
	var uploadedChunks []model.SysFileChunk
	if err := db.Where("upload_id = ?", req.UploadID).Find(&uploadedChunks).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询分片记录失败")
	}

	if len(uploadedChunks) != session.ChunkCount {
		return nil, errors.New("分片未全部上传")
	}

	// 创建最终文件存储目录
	now := time.Now()
	storeDir := filepath.Join("uploads", "files", fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()))
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("创建存储目录失败")
	}

	// 生成最终文件名
	ext := strings.ToLower(filepath.Ext(session.FileName))
	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), generateRandomString(8), ext)
	finalPath := filepath.Join(storeDir, uniqueName)

	// 创建最终文件
	finalFile, err := os.Create(finalPath)
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("创建最终文件失败")
	}
	defer finalFile.Close()

	// 按顺序合并分片
	for i := 0; i < session.ChunkCount; i++ {
		chunkPath := filepath.Join("uploads", "chunks", fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()), req.UploadID, fmt.Sprintf("part_%d", i))
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			ctx.Log.Errorf("%+v", errors.WithStack(err))
			return nil, errors.New("读取分片文件失败")
		}
		if _, err := finalFile.Write(chunkData); err != nil {
			ctx.Log.Errorf("%+v", errors.WithStack(err))
			return nil, errors.New("写入最终文件失败")
		}
	}

	// 验证文件哈希（可选）
	hasher := sha256.New()
	finalFile.Seek(0, 0)
	if _, err := io.Copy(hasher, finalFile); err != nil {
		ctx.Log.Warnf("计算文件哈希失败: %v", err)
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))
	ctx.Log.Infof("文件哈希: %s", fileHash)

	// 保存文件记录到数据库
	sysFile := model.SysFile{
		FileName:     session.FileName,
		FilePath:     finalPath,
		FileSize:     session.FileSize,
		FileExt:      ext,
		FolderID:     session.ParentFolderID,
		UploadUserID: session.UploadUserID,
	}

	if err := db.Create(&sysFile).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("保存文件记录失败")
	}

	// 清理分片文件（可选，保留以支持断点续传）
	// 可以添加定时任务清理旧的分片文件

	// 删除会话信息
	delete(sessionStore, req.UploadID)

	// 返回完整URL
	fileURL := config.ServerConf.BaseUrl + "/" + finalPath

	return &types.ChunkMergeResponse{
		AssetID:  sysFile.ID,
		FileName: sysFile.FileName,
		FileURL:  fileURL,
		FileSize: sysFile.FileSize,
	}, nil
}

// GetChunkStatus 获取分片上传状态（断点续传）
func GetChunkStatus(ctx *svc.ServiceContext, uploadID string) (*types.ChunkStatusResponse, error) {
	// 获取会话信息
	session, ok := sessionStore[uploadID]
	if !ok {
		return nil, errors.New("上传会话不存在或已过期")
	}

	db := model.DB()

	// 查询已上传的分片
	var chunks []model.SysFileChunk
	if err := db.Where("upload_id = ?", uploadID).Find(&chunks).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询分片状态失败")
	}

	uploadedIndexes := make([]int, 0, len(chunks))
	for _, chunk := range chunks {
		uploadedIndexes = append(uploadedIndexes, chunk.ChunkIndex)
	}

	return &types.ChunkStatusResponse{
		UploadID:       uploadID,
		UploadedChunks: uploadedIndexes,
		TotalChunks:    session.ChunkCount,
	}, nil
}
