// Command import 把「视频元数据(+字幕)」从 JSON 一键入库。
// 用法：go run ./cmd/import -config configs/config.yaml -file docs/samples/video.import.json
// JSON 支持单个对象或对象数组。
// 字幕可二选一：直接给已传 CDN 的 url（subtitle_*_url），或给本地路径（subtitle_*_file）
// 由本命令上传 COS 后回填 url（需在 .env 配好 COS 密钥，且事先轮换泄露密钥）。
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/echotalk/echotalk_server/internal/bootstrap"
	"github.com/echotalk/echotalk_server/internal/config"
	"github.com/echotalk/echotalk_server/internal/module/content"
	"github.com/echotalk/echotalk_server/internal/pkg/cos"
)

type importVideo struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	CoverURL      string `json:"cover_url"`
	HLSURL        string `json:"hls_url"`
	SubtitleEnURL string `json:"subtitle_en_url"`
	SubtitleCnURL string `json:"subtitle_cn_url"`
	// 本地字幕文件路径（可选）：填了则上传 COS 并回填上面对应的 *_url。
	SubtitleEnFile string `json:"subtitle_en_file"`
	SubtitleCnFile string `json:"subtitle_cn_file"`
	Duration       int    `json:"duration"`
	Difficulty     int8   `json:"difficulty"`
	Category       string `json:"category"`
	IsFree         bool   `json:"is_free"`
	Sort           int    `json:"sort"`
	Status         *int8  `json:"status"` // 默认上架(1)
}

// needsUpload 是否有任一条目带本地字幕路径，需要构建 COS 上传器。
func needsUpload(items []importVideo) bool {
	for _, it := range items {
		if it.SubtitleEnFile != "" || it.SubtitleCnFile != "" {
			return true
		}
	}
	return false
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
	// 兜底建表：允许在未先启动 server 的全新库上直接导入。
	if err := bootstrap.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate: %v", err)
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

	// 仅当有本地字幕需要上传时才构建 COS 上传器（纯 URL 导入无需 COS 凭据）。
	var uploader *cos.Uploader
	if needsUpload(items) {
		uploader, err = cos.New(cfg.COS)
		if err != nil {
			log.Fatalf("init cos: %v", err)
		}
	}

	ctx := context.Background()
	repo := content.NewRepository(db)
	var ok, failed int
	for i, it := range items {
		if err := importOne(ctx, repo, uploader, it); err != nil {
			failed++
			log.Printf("import #%d (%q) failed: %v", i+1, it.Title, err)
			continue
		}
		ok++
	}
	log.Printf("done: %d ok, %d failed (total %d)", ok, failed, len(items))
	if failed > 0 {
		os.Exit(1)
	}
}

// importOne 处理单条：必要时上传本地字幕，再写库。
func importOne(ctx context.Context, repo *content.Repository, uploader *cos.Uploader, it importVideo) error {
	enURL, err := resolveSubtitle(ctx, uploader, it.SubtitleEnURL, it.SubtitleEnFile)
	if err != nil {
		return err
	}
	cnURL, err := resolveSubtitle(ctx, uploader, it.SubtitleCnURL, it.SubtitleCnFile)
	if err != nil {
		return err
	}

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
		HLSURL: it.HLSURL, SubtitleEnURL: enURL, SubtitleCnURL: cnURL,
		Duration: it.Duration, Difficulty: diff, Category: it.Category,
		IsFree: it.IsFree, Status: status, Sort: it.Sort,
	}
	if err := repo.Create(v); err != nil {
		return err
	}
	log.Printf("imported: id=%d title=%q is_free=%v status=%d en=%q", v.ID, v.Title, v.IsFree, v.Status, enURL)
	return nil
}

// resolveSubtitle 给定 url 与本地文件路径：有本地文件则上传 COS 取 url，否则用已有 url。
func resolveSubtitle(ctx context.Context, uploader *cos.Uploader, url, localFile string) (string, error) {
	if localFile == "" {
		return url, nil
	}
	if uploader == nil {
		return "", fmt.Errorf("提供了本地字幕 %q 但 COS 未初始化", localFile)
	}
	key := fmt.Sprintf("subtitle/%d_%s", time.Now().UnixNano(), filepath.Base(localFile))
	return uploader.UploadFile(ctx, localFile, key)
}
