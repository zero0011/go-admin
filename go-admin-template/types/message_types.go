package types

// ========== 用户侧消息 ==========

// MessageListRequest 消息列表请求
type MessageListRequest struct {
	Page     int  `form:"page" label:"页码"`
	PageSize int  `form:"page_size" label:"每页条数"`
	IsRead   *int `form:"is_read" label:"已读状态 0:未读 1:已读"`
}

// MessageListItem 消息列表项
type MessageListItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	PublishTime string `json:"publish_time"` // 格式 2006-01-02 15:04:05
	IsRead      bool   `json:"is_read"`
}

// MessageListResponse 消息列表响应
type MessageListResponse struct {
	Total int64             `json:"total"`
	List  []MessageListItem `json:"list"`
}

// MessageUnreadCountResponse 未读数响应
type MessageUnreadCountResponse struct {
	Count int64 `json:"count"`
}

// MessageReadRequest 标记已读请求
type MessageReadRequest struct {
	ID int `json:"id" binding:"required,min=1" label:"消息ID"`
}

// MessageReadResponse 标记已读响应
type MessageReadResponse struct {
	Count int64 `json:"count"` // 本次标记数量
}

// ========== 运营侧公告 ==========

// MessageAnnouncementCreateRequest 发布公告请求
type MessageAnnouncementCreateRequest struct {
	Title   string `json:"title" binding:"required,max=255" label:"标题"`
	Content string `json:"content" binding:"required,max=5000" label:"正文"`
}

// MessageAnnouncementListItem 已发布公告列表项
type MessageAnnouncementListItem struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	PublisherID   int    `json:"publisher_id"`
	PublisherName string `json:"publisher_name"`
	PublishTime   string `json:"publish_time"` // 格式 2006-01-02 15:04:05
}

// MessageAnnouncementListResponse 已发布公告列表响应
type MessageAnnouncementListResponse struct {
	Total int64                         `json:"total"`
	List  []MessageAnnouncementListItem `json:"list"`
}

// MessageAnnouncementListRequest 已发布公告列表请求
type MessageAnnouncementListRequest struct {
	Page     int `form:"page" label:"页码"`
	PageSize int `form:"page_size" label:"每页条数"`
}
