package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameSysMessage = "sys_message"

// SysMessage 站内信公告
type SysMessage struct {
	ID          int            `gorm:"column:id;primaryKey;autoIncrement:true"`
	Title       string         `gorm:"column:title;size:255;not null;comment:标题"`
	Content     string         `gorm:"column:content;type:text;not null;comment:正文(纯文本)"`
	PublisherID int            `gorm:"column:publisher_id;not null;comment:发布人ID"`
	PublishTime time.Time      `gorm:"column:publish_time;not null;comment:发布时间"`
	CreateTime  time.Time      `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time      `gorm:"column:update_time;autoUpdateTime"`
	DeleteTime  gorm.DeletedAt `gorm:"column:delete_time"`
}

func (*SysMessage) TableName() string {
	return TableNameSysMessage
}

const TableNameSysMessageRead = "sys_message_read"

// SysMessageRead 站内信已读记录
type SysMessageRead struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement:true"`
	MessageID int       `gorm:"column:message_id;not null;uniqueIndex:uk_message_user;comment:消息ID"`
	UserID    int       `gorm:"column:user_id;not null;uniqueIndex:uk_message_user;comment:用户ID"`
	ReadTime  time.Time `gorm:"column:read_time;not null;comment:已读时间"`
}

func (*SysMessageRead) TableName() string {
	return TableNameSysMessageRead
}
