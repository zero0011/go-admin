package message

import (
	"time"

	"go-admin-template/model"
	"go-admin-template/pkg/sse"
	"go-admin-template/svc"
	"go-admin-template/types"

	"github.com/pkg/errors"
)

// CreateAnnouncement 创建并发布公告（写库成功后广播实时事件）
func CreateAnnouncement(ctx *svc.ServiceContext, req *types.MessageAnnouncementCreateRequest) error {
	userID := getUserID(ctx)
	msg := model.SysMessage{
		Title:       req.Title,
		Content:     req.Content,
		PublisherID: userID,
		PublishTime: time.Now(),
	}
	if err := model.DB().Create(&msg).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return errors.New("发布公告失败")
	}

	// 广播失败仅记录日志，不影响发布结果；离线/未送达用户由列表与未读数接口补齐
	if err := sse.PublishNewMessage(sse.MessagePayload{
		ID:          msg.ID,
		Title:       msg.Title,
		PublishTime: msg.PublishTime.Format(timeFormat),
	}); err != nil {
		ctx.Log.Errorf("sse publish failed: %+v", errors.WithStack(err))
	}
	return nil
}

// AnnouncementList 已发布公告列表（分页，含发布人姓名）
func AnnouncementList(ctx *svc.ServiceContext, req *types.MessageAnnouncementListRequest) (*types.MessageAnnouncementListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	db := model.DB()

	var total int64
	if err := db.Table("sys_message").Count(&total).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询公告总数失败")
	}

	type row struct {
		model.SysMessage
		PublisherName string `gorm:"column:publisher_name"`
	}

	var rows []row
	if err := db.Table("sys_message").
		Select("sys_message.*, admin_user.realname AS publisher_name").
		Joins("LEFT JOIN admin_user ON admin_user.id = sys_message.publisher_id AND admin_user.delete_time IS NULL").
		Order("sys_message.publish_time DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&rows).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询公告列表失败")
	}

	list := make([]types.MessageAnnouncementListItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.MessageAnnouncementListItem{
			ID:            r.ID,
			Title:         r.Title,
			Content:       r.Content,
			PublisherID:   r.PublisherID,
			PublisherName: r.PublisherName,
			PublishTime:   r.PublishTime.Format(timeFormat),
		})
	}
	return &types.MessageAnnouncementListResponse{Total: total, List: list}, nil
}
