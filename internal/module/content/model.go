package content

import (
	"time"

	"gorm.io/gorm"
)

// 视频上下架状态。
const (
	VideoStatusDraft   int8 = 0 // 草稿
	VideoStatusOnline  int8 = 1 // 上架
	VideoStatusOffline int8 = 2 // 下架
)

// Video 视频内容实体。句级数据改由字幕文件承载（见 SubtitleEnURL / SubtitleCnURL）。
type Video struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Title         string         `gorm:"size:128" json:"title"`              // 标题
	Description   string         `gorm:"size:512" json:"description"`        // 简介
	CoverURL      string         `gorm:"size:255" json:"cover_url"`          // 封面地址(CDN)
	HLSURL        string         `gorm:"size:255" json:"hls_url"`            // HLS播放地址(CDN)
	SubtitleEnURL string         `gorm:"size:255" json:"subtitle_en_url"`    // 英文字幕文件地址(含句级时间戳)
	SubtitleCnURL string         `gorm:"size:255" json:"subtitle_cn_url"`    // 中文字幕文件地址
	Duration      int            `json:"duration"`                           // 时长(秒)
	Difficulty    int8           `gorm:"default:1" json:"difficulty"`        // 难度 1-5
	Category      string         `gorm:"size:32;index" json:"category"`      // 分类
	IsFree        bool           `gorm:"default:false;index" json:"is_free"` // 是否免费(付费边界)
	Status        int8           `gorm:"default:0;index" json:"status"`      // 上下架状态
	Sort          int            `gorm:"default:0" json:"sort"`              // 排序权重
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// TableName 指定表名。
func (Video) TableName() string { return "videos" }
