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
	codePrefix := ""
	if len(code) > 10 {
		codePrefix = code[:10] + "..."
	} else {
		codePrefix = code
	}
	secretPrefix := ""
	if len(w.appSecret) > 10 {
		secretPrefix = w.appSecret[:10] + "..."
	} else {
		secretPrefix = w.appSecret
	}
	w.log.Infof("Calling WeChat API: appid=%s, code_length=%d, code_prefix=%s, secret_length=%d, secret_prefix=%s",
		w.appID, len(code), codePrefix, len(w.appSecret), secretPrefix)

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
		var suggestion string
		switch result.ErrCode {
		case 40029:
			errMsg = "code 无效或已过期（code 只能使用一次，且有时效性）"
			suggestion = "请检查：1) code 是否被重复使用 2) code 是否过期（约5分钟）3) AppID/AppSecret 是否正确"
		case 40163:
			errMsg = "code 已被使用（每个 code 只能使用一次）"
			suggestion = "请重新获取 code 并立即使用"
		case 40013:
			errMsg = "AppID 无效，请检查配置"
			suggestion = "请确认配置的 app_id 与微信公众平台中的 AppID 完全一致"
		case 40125:
			errMsg = "AppSecret 无效，请检查配置"
			suggestion = "请确认配置的 app_secret 与微信公众平台中的 AppSecret 完全一致"
		default:
			errMsg = result.ErrMsg
			suggestion = "请检查微信 API 文档了解错误详情"
		}
		w.log.Errorf("WeChat API error: code=%d, msg=%s, detail=%s, suggestion=%s, appid=%s",
			result.ErrCode, result.ErrMsg, errMsg, suggestion, w.appID)
		return "", "", fmt.Errorf("wechat api error [%d]: %s. %s", result.ErrCode, errMsg, suggestion)
	}

	if result.OpenID == "" {
		return "", "", fmt.Errorf("openid is empty in response")
	}

	return result.OpenID, result.SessionKey, nil
}
