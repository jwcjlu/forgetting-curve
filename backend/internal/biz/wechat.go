package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"forgetting-curve/backend/internal/conf"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// WechatService 微信服务接口
type WechatService interface {
	Code2Session(ctx context.Context, code string) (openid, sessionKey string, err error)
}

type wechatService struct {
	appID     string
	appSecret string
	client    *http.Client
	log       *log.Helper
}

// NewWechatService 创建微信服务
func NewWechatService(ac *conf.Wechat, logger log.Logger) WechatService {
	// 如果配置为 nil 或配置为空，返回 nil（service 层会处理）
	if ac == nil || ac.AppID == "" || ac.AppSecret == "" {
		log.NewHelper(logger).Warn("WeChat service not configured, AppID or AppSecret is empty")
		return nil
	}

	return &wechatService{
		appID:     ac.AppID,
		appSecret: ac.AppSecret,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log.NewHelper(logger),
	}
}

// Code2SessionResponse 微信 code2session 响应
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Code2Session 通过 code 换取 openid 和 session_key
func (w *wechatService) Code2Session(ctx context.Context, code string) (openid, sessionKey string, err error) {
	if w.appID == "" || w.appSecret == "" {
		return "", "", fmt.Errorf("wechat app_id or app_secret not configured")
	}

	if code == "" {
		return "", "", fmt.Errorf("code cannot be empty")
	}

	// 构建请求 URL
	apiURL := "https://api.weixin.qq.com/sns/jscode2session"
	params := url.Values{}
	params.Set("appid", w.appID)
	params.Set("secret", w.appSecret)
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")

	fullURL := apiURL + "?" + params.Encode()

	// 记录请求信息（不记录完整的 secret）
	w.log.Infof("Calling WeChat API: appid=%s, code_length=%d", w.appID, len(code))

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请求
	resp, err := w.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to call wechat api: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	w.log.Infof("WeChat API response: %s", string(body))

	// 解析响应
	var result Code2SessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查错误
	if result.ErrCode != 0 {
		// 根据错误码提供更详细的错误信息
		var errMsg string
		switch result.ErrCode {
		case 40029:
			errMsg = "code 无效或已过期（code 只能使用一次，且有时效性）"
		case 40163:
			errMsg = "code 已被使用（每个 code 只能使用一次）"
		case 40013:
			errMsg = "AppID 无效，请检查配置"
		case 40125:
			errMsg = "AppSecret 无效，请检查配置"
		default:
			errMsg = result.ErrMsg
		}
		w.log.Errorf("WeChat API error: code=%d, msg=%s, detail=%s", result.ErrCode, result.ErrMsg, errMsg)
		return "", "", fmt.Errorf("wechat api error [%d]: %s", result.ErrCode, errMsg)
	}

	if result.OpenID == "" {
		return "", "", fmt.Errorf("openid is empty in response")
	}

	return result.OpenID, result.SessionKey, nil
}
