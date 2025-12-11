package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"forgetting-curve/backend/internal/biz"
	"io"
	"net/http"
	"strings"
	"time"
)

// UmiOCRClient HTTP客户端
type UmiOCRClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewUmiOCRClient 创建HTTP客户端
func NewUmiOCRClient() biz.OCRService {
	return &UmiOCRClient{
		BaseURL: "http://127.0.0.1:1224",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
}

// OCRFromBase64 Base64图片识别
// https://github.com/hiroi-sora/Umi-OCR/blob/main/docs/http/api_ocr.md#/api/ocr
func (c *UmiOCRClient) RecognizeText(ctx context.Context, ocrReq *biz.OCRRequest) (*biz.OCRResult, error) {
	// 构建请求数据
	requestData := map[string]interface{}{
		"base64": ocrReq.Base64,
	}

	// 添加可选参数
	if ocrReq.Options != nil && len(ocrReq.Options) > 0 {
		requestData["options"] = ocrReq.Options
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/ocr", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result biz.OCRResult
	// 先尝试直接解析
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		// 如果解析失败，可能是转义换行符问题（文档中提到的问题）
		// 某些HTTP库可能会自动将转义换行符\\n转换为真实换行\n
		// 尝试将真实换行替换回转义换行
		bodyStr := string(bodyBytes)
		bodyStr = strings.ReplaceAll(bodyStr, "\n", "\\n")
		bodyStr = strings.ReplaceAll(bodyStr, "\r", "\\r")

		if err2 := json.Unmarshal([]byte(bodyStr), &result); err2 != nil {
			return nil, fmt.Errorf("解析响应失败: %v (原始错误: %v), 响应内容: %s", err2, err, string(bodyBytes))
		}
	}

	return &result, nil
}
