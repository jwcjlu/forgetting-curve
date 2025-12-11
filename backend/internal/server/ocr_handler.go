package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"forgetting-curve/backend/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// OCRRequest OCR 请求结构

// OCRResponse OCR 响应结构（符合 Umi-OCR API 格式）
type OCRResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg,omitempty"`
}

// OCRHandler OCR HTTP 处理器
type OCRHandler struct {
	ocrService biz.OCRService
	log        *log.Helper
}

// NewOCRHandler 创建 OCR 处理器
func NewOCRHandler(ocrService biz.OCRService, logger log.Logger) *OCRHandler {
	return &OCRHandler{
		ocrService: ocrService,
		log:        log.NewHelper(logger),
	}
}

// HandleOCR 处理 OCR 请求
func (h *OCRHandler) HandleOCR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, 400, "Method not allowed, only POST is supported")
		return
	}

	// 解析请求
	var req biz.OCRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorf("Failed to decode request: %v", err)
		h.writeError(w, 400, "Invalid request body")
		return
	}

	// 验证 base64 数据
	if req.Base64 == "" {
		h.writeError(w, 400, "base64 field is required")
		return
	}

	// 获取数据格式选项（默认为 dict）
	dataFormat := "dict"
	if req.Options != nil {
		if format, ok := req.Options["data.format"].(string); ok {
			dataFormat = format
		}
	}

	// 调用 OCR 服务
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	resp, err := h.ocrService.RecognizeText(ctx, &req)
	if err != nil {
		h.log.Errorf("OCR recognition failed: %v", err)
		// 使用 900+ 错误码符合 Umi-OCR API 规范
		h.writeError(w, 900, err.Error())
		return
	}

	// 处理识别结果
	if resp.Code == 0 {
		// 无文本
		h.writeResponse(w, 101, "图片中未识别到文本", dataFormat)
		return
	}
	result := resp.Data
	h.writeResponse(w, 100, "", result)

}

// writeResponse 写入成功响应
func (h *OCRHandler) writeResponse(w http.ResponseWriter, code int, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := OCRResponse{
		Code: code,
		Data: data,
	}
	if msg != "" {
		response.Msg = msg
	}

	// 使用 json.Marshal 确保 Unicode 转义（符合 Umi-OCR API 规范）
	jsonData, err := json.Marshal(response)
	if err != nil {
		h.log.Errorf("Failed to marshal response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonData)
}

// writeError 写入错误响应
func (h *OCRHandler) writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := OCRResponse{
		Code: code,
		Data: msg,
		Msg:  msg,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		h.log.Errorf("Failed to marshal error response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonData)
}
