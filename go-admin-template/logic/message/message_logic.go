package message

import (
	"strings"
	"time"

	"go-admin-template/model"
	"go-admin-template/pkg/jwt"
	"go-admin-template/svc"
	"go-admin-template/types"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// timeFormat 统一的时间输出格式
const timeFormat = "2006-01-02 15:04:05"

// getUserID 从上下文中获取当前用户ID
func getUserID(ctx *svc.ServiceContext) int {
	user, _ := ctx.GinContext.Get(jwt.UserInfo)
	if claims, ok := user.(*jwt.Claims); ok {
		return claims.Id
	}
	return 0
}

// ListMessages 分页消息列表（含已读标记，可按已读/未读筛选）
func ListMessages(ctx *svc.ServiceContext, req *types.MessageListRequest) (*types.MessageListResponse, error) {
	userID := getUserID(ctx)
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
	applyFilter := func(q *gorm.DB) *gorm.DB {
		sub := db.Model(&model.SysMessageRead{}).Select("message_id").Where("user_id = ?", userID)
		switch {
		case req.IsRead != nil && *req.IsRead == 1:
			return q.Where("id IN (?)", sub)
		case req.IsRead != nil && *req.IsRead == 0:
			return q.Where("id NOT IN (?)", sub)
		}
		return q
	}

	var total int64
	if err := applyFilter(db.Model(&model.SysMessage{})).Count(&total).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询消息总数失败")
	}

	var messages []model.SysMessage
	if err := applyFilter(db.Model(&model.SysMessage{})).
		Order("publish_time DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&messages).Error; err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询消息列表失败")
	}

	// 本页消息的已读集合
	readSet := make(map[int]bool, len(messages))
	if len(messages) > 0 {
		ids := make([]int, 0, len(messages))
		for _, m := range messages {
			ids = append(ids, m.ID)
		}
		var reads []model.SysMessageRead
		if err := db.Where("user_id = ? AND message_id IN ?", userID, ids).Find(&reads).Error; err != nil {
			ctx.Log.Errorf("%+v", errors.WithStack(err))
			return nil, errors.New("查询已读记录失败")
		}
		for _, r := range reads {
			readSet[r.MessageID] = true
		}
	}

	list := make([]types.MessageListItem, 0, len(messages))
	for _, m := range messages {
		list = append(list, types.MessageListItem{
			ID:          m.ID,
			Title:       m.Title,
			Content:     m.Content,
			PublishTime: m.PublishTime.Format(timeFormat),
			IsRead:      readSet[m.ID],
		})
	}
	return &types.MessageListResponse{Total: total, List: list}, nil
}

// UnreadCount 未读数（当前用户）
func UnreadCount(ctx *svc.ServiceContext) (*types.MessageUnreadCountResponse, error) {
	return UnreadCountForUser(getUserID(ctx))
}

// UnreadCountForUser 指定用户的未读数 = 已发布消息总数 - 该用户已读记录数
// 仅全体广播，无需按目标人群过滤
func UnreadCountForUser(userID int) (*types.MessageUnreadCountResponse, error) {
	db := model.DB()
	var published, read int64
	if err := db.Model(&model.SysMessage{}).Count(&published).Error; err != nil {
		return nil, errors.WithStack(errors.New("查询消息总数失败"))
	}
	if err := db.Model(&model.SysMessageRead{}).Where("user_id = ?", userID).Count(&read).Error; err != nil {
		return nil, errors.WithStack(errors.New("查询已读记录失败"))
	}
	count := published - read
	if count < 0 {
		count = 0
	}
	return &types.MessageUnreadCountResponse{Count: count}, nil
}

// MarkRead 单条已读（幂等，返回本次标记数量）
func MarkRead(ctx *svc.ServiceContext, req *types.MessageReadRequest) (*types.MessageReadResponse, error) {
	userID := getUserID(ctx)
	db := model.DB()

	var msg model.SysMessage
	if err := db.First(&msg, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("消息不存在")
		}
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询消息失败")
	}

	res := db.Exec(
		"INSERT IGNORE INTO sys_message_read (message_id, user_id, read_time) VALUES (?, ?, ?)",
		req.ID, userID, time.Now(),
	)
	if res.Error != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(res.Error))
		return nil, errors.New("标记已读失败")
	}
	return &types.MessageReadResponse{Count: res.RowsAffected}, nil
}

// MarkAllRead 一键已读（幂等，返回本次标记数量）
func MarkAllRead(ctx *svc.ServiceContext) (*types.MessageReadResponse, error) {
	userID := getUserID(ctx)
	db := model.DB()

	var unreadIDs []int
	err := db.Model(&model.SysMessage{}).
		Where("id NOT IN (?)", db.Model(&model.SysMessageRead{}).Select("message_id").Where("user_id = ?", userID)).
		Pluck("id", &unreadIDs).Error
	if err != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(err))
		return nil, errors.New("查询未读消息失败")
	}
	if len(unreadIDs) == 0 {
		return &types.MessageReadResponse{Count: 0}, nil
	}

	now := time.Now()
	valueStrings := make([]string, 0, len(unreadIDs))
	valueArgs := make([]any, 0, len(unreadIDs)*3)
	for _, id := range unreadIDs {
		valueStrings = append(valueStrings, "(?, ?, ?)")
		valueArgs = append(valueArgs, id, userID, now)
	}
	stmt := "INSERT IGNORE INTO sys_message_read (message_id, user_id, read_time) VALUES " + strings.Join(valueStrings, ",")
	res := db.Exec(stmt, valueArgs...)
	if res.Error != nil {
		ctx.Log.Errorf("%+v", errors.WithStack(res.Error))
		return nil, errors.New("标记已读失败")
	}
	return &types.MessageReadResponse{Count: res.RowsAffected}, nil
}
