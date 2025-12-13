package biz

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"forgetting-curve/backend/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
)

func TestNewLLMService(t *testing.T) {
	tests := []struct {
		name    string
		config  *conf.LLM
		wantNil bool
	}{
		{
			name:    "配置为空",
			config:  nil,
			wantNil: true,
		},
		{
			name: "API Key为空",
			config: &conf.LLM{
				APIKey: "",
			},
			wantNil: true,
		},
		{
			name: "正常配置",
			config: &conf.LLM{
				APIKey:  "test-api-key",
				BaseURL: "https://api.test.com/v1",
				Model:   "gpt-4",
			},
			wantNil: false,
		},
		{
			name: "使用默认值",
			config: &conf.LLM{
				APIKey: "test-api-key",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用一个有效的logger，避免nil指针
			// 使用io.Discard来丢弃日志输出
			logger := log.NewStdLogger(io.Discard)
			service := NewLLMService(tt.config, logger)

			if tt.wantNil && service != nil {
				t.Errorf("NewLLMService() = %v, want nil", service)
			}
			if !tt.wantNil && service == nil {
				t.Errorf("NewLLMService() = nil, want non-nil")
			}
		})
	}
}

func TestGenerateReviewQuestions_Success(t *testing.T) {
	// 创建模拟的HTTP服务器
	mockResponse := OpenAIResponse{
		Choices: []Choice{
			{
				Message: Message{
					Role:    "assistant",
					Content: `{"questions":[{"type":"multiple_choice","question":"哪个单词的意思是'你好'？","options":["hello","world","test","good"],"correct_answer":"hello"},{"type":"fill_blank","question":"请填入单词：____ means hello in Chinese.","correct_answer":"hello"}]}`,
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		// 验证Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got %s", authHeader)
		}

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// 创建LLM服务
	config := &conf.LLM{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Model:   "gpt-3.5-turbo",
	}
	// 使用一个有效的logger，避免nil指针
	// 使用io.Discard来丢弃日志输出
	logger := log.NewStdLogger(io.Discard)
	service := NewLLMService(config, logger)

	// 如果service为nil，说明创建失败
	if service == nil {
		t.Fatal("NewLLMService() returned nil")
	}

	ctx := context.Background()
	questions, err := service.GenerateReviewQuestions(ctx, "hello", "你好", "初中一年级")

	if err != nil {
		t.Fatalf("GenerateReviewQuestions() error = %v, want nil", err)
	}

	if len(questions) != 2 {
		t.Fatalf("GenerateReviewQuestions() returned %d questions, want 2", len(questions))
	}

	// 验证第一道题（选择题）
	if questions[0].Type != "multiple_choice" {
		t.Errorf("Question 0 type = %s, want multiple_choice", questions[0].Type)
	}
	if len(questions[0].Options) != 4 {
		t.Errorf("Question 0 options count = %d, want 4", len(questions[0].Options))
	}
	if questions[0].CorrectAnswer != "hello" {
		t.Errorf("Question 0 correct answer = %s, want hello", questions[0].CorrectAnswer)
	}

	// 验证第二道题（填空题）
	if questions[1].Type != "fill_blank" {
		t.Errorf("Question 1 type = %s, want fill_blank", questions[1].Type)
	}
	if questions[1].CorrectAnswer != "hello" {
		t.Errorf("Question 1 correct answer = %s, want hello", questions[1].CorrectAnswer)
	}
}

func TestGenerateReviewQuestions_ErrorResponse(t *testing.T) {
	// 创建返回错误的模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := OpenAIResponse{
			Error: &Error{
				Message: "Invalid API key",
				Type:    "invalid_request_error",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer server.Close()

	config := &conf.LLM{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
		Model:   "gpt-3.5-turbo",
	}
	logger := log.NewStdLogger(io.Discard)
	service := NewLLMService(config, logger)

	ctx := context.Background()
	_, err := service.GenerateReviewQuestions(ctx, "hello", "你好", "")

	if err == nil {
		t.Error("GenerateReviewQuestions() error = nil, want error")
	}
}

func TestGenerateReviewQuestions_HTTPError(t *testing.T) {
	// 创建返回HTTP错误的模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	config := &conf.LLM{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-3.5-turbo",
	}
	logger := log.NewStdLogger(io.Discard)
	service := NewLLMService(config, logger)

	ctx := context.Background()
	_, err := service.GenerateReviewQuestions(ctx, "hello", "你好", "")

	if err == nil {
		t.Error("GenerateReviewQuestions() error = nil, want error")
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "纯JSON",
			input:    `{"questions":[]}`,
			expected: `{"questions":[]}`,
		},
		{
			name:     "带markdown代码块",
			input:    "```json\n{\"questions\":[]}\n```",
			expected: `{"questions":[]}`, // extractJSON应该能处理markdown代码块
		},
		{
			name:     "带其他文本",
			input:    "Here is the JSON:\n{\"questions\":[]}\nEnd",
			expected: `{"questions":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSON() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetGradeDifficultyHint(t *testing.T) {
	tests := []struct {
		name     string
		grade    string
		contains string
	}{
		{
			name:     "小学",
			grade:    "小学一年级",
			contains: "小学生",
		},
		{
			name:     "初中",
			grade:    "初中二年级",
			contains: "初中生",
		},
		{
			name:     "高中",
			grade:    "高中三年级",
			contains: "高中生",
		},
		{
			name:     "大学",
			grade:    "大学",
			contains: "较高",
		},
		{
			name:     "其他",
			grade:    "其他",
			contains: "适中",
		},
		{
			name:     "空字符串",
			grade:    "",
			contains: "适中",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getGradeDifficultyHint(tt.grade)
			if !contains(result, tt.contains) {
				t.Errorf("getGradeDifficultyHint(%s) = %s, should contain %s", tt.grade, result, tt.contains)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "包含子串",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "不包含子串",
			s:        "hello world",
			substr:   "test",
			expected: false,
		},
		{
			name:     "空子串",
			s:        "hello",
			substr:   "",
			expected: true,
		},
		{
			name:     "完全匹配",
			s:        "hello",
			substr:   "hello",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%s, %s) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}
