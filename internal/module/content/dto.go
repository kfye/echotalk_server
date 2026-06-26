package content

// ListQuery App 列表查询参数。
type ListQuery struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Category   string `form:"category"`
	Difficulty int8   `form:"difficulty"`
}

// VideoListItem 列表项（不含媒体地址）。
type VideoListItem struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	CoverURL   string `json:"cover_url"`
	Duration   int    `json:"duration"`
	Difficulty int8   `json:"difficulty"`
	Category   string `json:"category"`
	IsFree     bool   `json:"is_free"`
	Locked     bool   `json:"locked"` // 服务端计算：当前用户是否被付费墙拦住
}

// VideoDetail 详情。locked=true 时媒体地址(hls_url/字幕)被隐藏为空。
type VideoDetail struct {
	VideoListItem
	Description   string `json:"description"`
	Status        int8   `json:"status"`
	HLSURL        string `json:"hls_url"`
	SubtitleEnURL string `json:"subtitle_en_url"`
	SubtitleCnURL string `json:"subtitle_cn_url"`
}

// VideoInput 管理端新增/编辑入参。
type VideoInput struct {
	Title         string `json:"title" binding:"required,max=128"`
	Description   string `json:"description" binding:"max=512"`
	CoverURL      string `json:"cover_url" binding:"max=255"`
	HLSURL        string `json:"hls_url" binding:"max=255"`
	SubtitleEnURL string `json:"subtitle_en_url" binding:"max=255"`
	SubtitleCnURL string `json:"subtitle_cn_url" binding:"max=255"`
	Duration      int    `json:"duration"`
	Difficulty    int8   `json:"difficulty" binding:"omitempty,min=1,max=5"`
	Category      string `json:"category" binding:"max=32"`
	IsFree        bool   `json:"is_free"`
	Sort          int    `json:"sort"`
}

// UpdateStatusInput 上下架 / 免费付费切换入参。
type UpdateStatusInput struct {
	Status *int8 `json:"status" binding:"omitempty,oneof=0 1 2"`
	IsFree *bool `json:"is_free"`
}

// toListItem 把模型转列表项，locked 由调用方算。
func toListItem(v *Video, locked bool) VideoListItem {
	return VideoListItem{
		ID:         v.ID,
		Title:      v.Title,
		CoverURL:   v.CoverURL,
		Duration:   v.Duration,
		Difficulty: v.Difficulty,
		Category:   v.Category,
		IsFree:     v.IsFree,
		Locked:     locked,
	}
}

// toDetail 把模型转详情；locked=true 时隐藏媒体地址（服务端付费门禁）。
func toDetail(v *Video, locked bool) VideoDetail {
	d := VideoDetail{
		VideoListItem: toListItem(v, locked),
		Description:   v.Description,
		Status:        v.Status,
		HLSURL:        v.HLSURL,
		SubtitleEnURL: v.SubtitleEnURL,
		SubtitleCnURL: v.SubtitleCnURL,
	}
	if locked {
		d.HLSURL = ""
		d.SubtitleEnURL = ""
		d.SubtitleCnURL = ""
	}
	return d
}
