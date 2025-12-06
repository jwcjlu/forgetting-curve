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

	w.log.Infof("Calling WeChat API: %s", fullURL)

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
		return "", "", fmt.Errorf("wechat api error: code=%d, msg=%s", result.ErrCode, result.ErrMsg)
	}

	if result.OpenID == "" {
		return "", "", fmt.Errorf("openid is empty in response")
	}

	return result.OpenID, result.SessionKey, nil
}
