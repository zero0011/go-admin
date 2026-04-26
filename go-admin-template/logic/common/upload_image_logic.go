package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-admin-template/config"
	"go-admin-template/svc"
	"go-admin-template/types"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// UploadImage 上传图片/视频
func UploadImage(ctx *svc.ServiceContext, req *types.UploadImageRequest) (resp string, err error) {
	ext := strings.ToLower(filepath.Ext(req.File.Filename))

	// 支持的图片和视频格式
	imageExts := []string{".png", ".jpg", ".jpeg", ".gif"}
	videoExts := []string{".mp4", ".mov", ".avi", ".webm"}
	allowedExts := append(imageExts, videoExts...)

	if !lo.Contains(allowedExts, ext) {
		err = errors.New("不支持的文件格式")
		return
	}

	// 图片限制 2MB，视频限制 100MB
	// maxSize := int64(2 * 1024 * 1024)
	// if lo.Contains(videoExts, ext) {
	// 	maxSize = 100 * 1024 * 1024
	// }
	// if req.File.Size > maxSize {
	// 	err = errors.New("文件大小超出限制")
	// 	return
	// }

	open, err := req.File.Open()
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		err = errors.New("系统错误")
		return
	}
	defer open.Close()

	// 生成文件名：时间戳+扩展名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	uploadDir := "uploads"

	// 确保上传目录存在
	if err = os.MkdirAll(uploadDir, 0755); err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		err = errors.New("系统错误")
		return
	}

	// 保存文件到本地
	savePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(savePath)
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		err = errors.New("系统错误")
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, open); err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		err = errors.New("系统错误")
		return
	}

	// 返回完整 URL
	resp = config.ServerConf.BaseUrl + "/" + savePath
	return
}
