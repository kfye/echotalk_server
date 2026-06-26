// Command import 把「视频元数据(+字幕 URL)」从 JSON 一键入库。
// 用法：go run ./cmd/import -config configs/config.yaml -file docs/samples/video.import.json
// JSON 支持单个对象或对象数组。媒体/字幕需先手动传 COS，这里只写 CDN URL 入库。
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/echotalk/echotalk_server/internal/bootstrap"
	"github.com/echotalk/echotalk_server/internal/config"
	"github.com/echotalk/echotalk_server/internal/module/content"
)

type importVideo struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	CoverURL      string `json:"cover_url"`
	HLSURL        string `json:"hls_url"`
	SubtitleEnURL string `json:"subtitle_en_url"`
	SubtitleCnURL string `json:"subtitle_cn_url"`
	Duration      int    `json:"duration"`
	Difficulty    int8   `json:"difficulty"`
	Category      string `json:"category"`
	IsFree        bool   `json:"is_free"`
	Sort          int    `json:"sort"`
	Status        *int8  `json:"status"` // 默认上架(1)
}

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file")
	file := flag.String("file", "", "video metadata JSON (object or array)")
	flag.Parse()
	if *file == "" {
		log.Fatal("missing -file")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := bootstrap.InitDB(cfg.MySQL)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	raw, err := os.ReadFile(*file)
	if err != nil {
		log.Fatalf("read file: %v", err)
	}
	// 兼容单对象或数组
	var items []importVideo
	if err := json.Unmarshal(raw, &items); err != nil {
		var one importVideo
		if err2 := json.Unmarshal(raw, &one); err2 != nil {
			log.Fatalf("parse json: %v", err)
		}
		items = []importVideo{one}
	}

	repo := content.NewRepository(db)
	for i, it := range items {
		status := content.VideoStatusOnline
		if it.Status != nil {
			status = *it.Status
		}
		diff := it.Difficulty
		if diff <= 0 {
			diff = 1
		}
		v := &content.Video{
			Title: it.Title, Description: it.Description, CoverURL: it.CoverURL,
			HLSURL: it.HLSURL, SubtitleEnURL: it.SubtitleEnURL, SubtitleCnURL: it.SubtitleCnURL,
			Duration: it.Duration, Difficulty: diff, Category: it.Category,
			IsFree: it.IsFree, Status: status, Sort: it.Sort,
		}
		if err := repo.Create(v); err != nil {
			log.Fatalf("import #%d failed: %v", i+1, err)
		}
		log.Printf("imported #%d: id=%d title=%q is_free=%v status=%d", i+1, v.ID, v.Title, v.IsFree, v.Status)
	}
	log.Printf("done, %d video(s) imported", len(items))
}
