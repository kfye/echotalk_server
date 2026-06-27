// Package cos 封装腾讯云 COS 上传，供导入命令等服务端场景使用。
// 用永久密钥直传，仅限离线工具/后端进程；密钥只来自环境变量（.env），不入仓库。
package cos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	cossdk "github.com/tencentyun/cos-go-sdk-v5"

	"github.com/echotalk/echotalk_server/internal/config"
)

// Uploader 封装一个面向单个存储桶的 COS 客户端。
type Uploader struct {
	client     *cossdk.Client
	bucketBase string // https://{bucket}.cos.{region}.myqcloud.com
	cdnBase    string // 可选 CDN 域名，空则回落 bucketBase
}

// New 用配置创建上传器；缺密钥/桶信息直接报错，提示去 .env 配置。
func New(cfg config.COSConfig) (*Uploader, error) {
	switch {
	case cfg.Bucket == "":
		return nil, errors.New("cos: bucket 为空（设置 cos.bucket）")
	case cfg.Region == "":
		return nil, errors.New("cos: region 为空（设置 cos.region）")
	case cfg.SecretID == "":
		return nil, errors.New("cos: secret_id 为空（设置环境变量 ECHOTALK_COS_SECRET_ID，勿写入仓库）")
	case cfg.SecretKey == "":
		return nil, errors.New("cos: secret_key 为空（设置环境变量 ECHOTALK_COS_SECRET_KEY，勿写入仓库）")
	}

	bucketBase := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region)
	u, err := url.Parse(bucketBase)
	if err != nil {
		return nil, fmt.Errorf("cos: 解析桶地址失败: %w", err)
	}
	client := cossdk.NewClient(&cossdk.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cossdk.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})
	return &Uploader{
		client:     client,
		bucketBase: bucketBase,
		cdnBase:    strings.TrimRight(cfg.CDNBase, "/"),
	}, nil
}

// UploadFile 把本地文件传到 objectKey，返回可访问 URL（优先 CDN 域名）。
func (u *Uploader) UploadFile(ctx context.Context, localPath, objectKey string) (string, error) {
	objectKey = strings.TrimLeft(objectKey, "/")
	if _, err := u.client.Object.PutFromFile(ctx, objectKey, localPath, nil); err != nil {
		return "", fmt.Errorf("cos: 上传 %s 失败: %w", localPath, err)
	}
	base := u.bucketBase
	if u.cdnBase != "" {
		base = u.cdnBase
	}
	return base + "/" + objectKey, nil
}
