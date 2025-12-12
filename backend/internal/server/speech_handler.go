package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"forgetting-curve/backend/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// SpeechRecognitionRequest 语音识别请求
type SpeechRecognitionRequest struct {
	AudioBase64  string `json:"audio_base64"`  // base64 编码的音频数据
	ExpectedWord string `json:"expected_word"` // 期望的单词
}

// SpeechRecognitionResponse 语音识别响应
type SpeechRecognitionResponse struct {
	Code           int    `json:"code"`
	Message        string `json:"message,omitempty"`
	RecognizedText string `json:"recognized_text,omitempty"`
	ExpectedWord   string `json:"expected_word,omitempty"`
	IsCorrect      bool   `json:"is_correct,omitempty"`
}

// SpeechHandler 语音识别 HTTP 处理器
type SpeechHandler struct {
	speechService biz.SpeechRecognitionService
	log           *log.Helper
}

// NewSpeechHandler 创建语音识别处理器
func NewSpeechHandler(speechService biz.SpeechRecognitionService, logger log.Logger) *SpeechHandler {
	return &SpeechHandler{
		speechService: speechService,
		log:           log.NewHelper(logger),
	}
}

// HandleSpeechRecognition 处理语音识别请求
func (h *SpeechHandler) HandleSpeechRecognition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, 400, "Method not allowed, only POST is supported")
		return
	}

	// 解析请求
	var req SpeechRecognitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorf("Failed to decode request: %v", err)
		h.writeError(w, 400, "Invalid request body")
		return
	}

	// 验证必填字段
	if req.AudioBase64 == "" {
		h.writeError(w, 400, "audio_base64 field is required")
		return
	}
	if req.ExpectedWord == "" {
		h.writeError(w, 400, "expected_word field is required")
		return
	}

	// 调用语音识别服务
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	recognizedText, isCorrect, err := h.speechService.RecognizeSpeech(ctx, req.AudioBase64, req.ExpectedWord)
	if err != nil {
		h.log.Errorf("Speech recognition failed: %v", err)
		h.writeError(w, 500, fmt.Sprintf("Speech recognition failed: %v", err))
		return
	}

	// 返回成功响应
	h.writeSuccess(w, recognizedText, req.ExpectedWord, isCorrect)
}

// writeSuccess 写入成功响应
func (h *SpeechHandler) writeSuccess(w http.ResponseWriter, recognizedText, expectedWord string, isCorrect bool) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := SpeechRecognitionResponse{
		Code:           200,
		Message:        "Success",
		RecognizedText: recognizedText,
		ExpectedWord:   expectedWord,
		IsCorrect:      isCorrect,
	}

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
func (h *SpeechHandler) writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := SpeechRecognitionResponse{
		Code:    code,
		Message: msg,
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
