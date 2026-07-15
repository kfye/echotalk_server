package iflytek

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// iseRequestLine ISE v2 WebSocket 请求行（鉴权签名固定用它）。
const iseRequestLine = "GET /v2/open-ise HTTP/1.1"

// buildAuthURL 生成讯飞 WebAPI 鉴权后的 wss 连接地址。
// 鉴权方案（讯飞通用）：对 "host/date/request-line" 三行做 HMAC-SHA256(APISecret) 签名，
// 组装 authorization 后随查询参数带上，host/date 也一并附上供服务端校验。
func buildAuthURL(host, apiKey, apiSecret string) string {
	date := time.Now().UTC().Format(http.TimeFormat) // RFC1123 GMT

	signatureOrigin := fmt.Sprintf("host: %s\ndate: %s\n%s", host, date, iseRequestLine)
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(signatureOrigin))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	authorizationOrigin := fmt.Sprintf(
		`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`,
		apiKey, signature,
	)
	authorization := base64.StdEncoding.EncodeToString([]byte(authorizationOrigin))

	q := url.Values{}
	q.Set("authorization", authorization)
	q.Set("date", date)
	q.Set("host", host)
	return fmt.Sprintf("wss://%s/v2/open-ise?%s", host, q.Encode())
}
