package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameFileFolder = "sys_file_folder"

// FileFolder 文件夹
type FileFolder struct {
	ID         int            `gorm:"column:id;primaryKey;autoIncrement:true"`
	FolderName string         `gorm:"column:folder_name;size:255;not null;comment:文件夹名称"`
	ParentID   *int           `gorm:"column:parent_id;index;comment:父文件夹ID"`
	CreatedBy  int            `gorm:"column:created_by;not null;comment:创建人"`
	CreateTime time.Time      `gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time      `gorm:"column:update_time;autoUpdateTime"`
	DeleteTime gorm.DeletedAt `gorm:"column:delete_time"`
}

// TableName FileFolder's table name
func (*FileFolder) TableName() string {
	return TableNameFileFolder
}

const TableNameSysFile = "sys_file"

// SysFile 文件
type SysFile struct {
	ID           int            `gorm:"column:id;primaryKey;autoIncrement:true"`
	FileName     string         `gorm:"column:file_name;size:255;not null;comment:文件名"`
	FilePath     string         `gorm:"column:file_path;size:500;not null;comment:文件存储路径"`
	FileSize     int64          `gorm:"column:file_size;not null;comment:文件大小(字节)"`
	FileExt      string         `gorm:"column:file_ext;size:20;comment:文件扩展名"`
	FolderID     *int           `gorm:"column:folder_id;index;comment:所属文件夹ID"`
	UploadUserID int            `gorm:"column:upload_user_id;not null;comment:上传用户ID"`
	CreateTime   time.Time      `gorm:"column:create_time;autoCreateTime"`
	UpdateTime   time.Time      `gorm:"column:update_time;autoUpdateTime"`
	DeleteTime   gorm.DeletedAt `gorm:"column:delete_time"`
}

// TableName SysFile's table name
func (*SysFile) TableName() string {
	return TableNameSysFile
}

const TableNameSysFileChunk = "sys_file_chunk"

// SysFileChunk 分片文件
type SysFileChunk struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement:true"`
	UploadID   string    `gorm:"column:upload_id;size:64;index;not null;comment:上传会话ID"`
	FileID     int       `gorm:"column:file_id;index;comment:关联文件ID"`
	ChunkIndex int       `gorm:"column:chunk_index;not null;comment:分片索引"`
	ChunkHash  string    `gorm:"column:chunk_hash;size:64;not null;comment:分片哈希"`
	ChunkPath  string    `gorm:"column:chunk_path;size:500;not null;comment:分片存储路径"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"`
}

// TableName SysFileChunk's table name
func (*SysFileChunk) TableName() string {
	return TableNameSysFileChunk
}
