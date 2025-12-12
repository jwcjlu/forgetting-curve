package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"forgetting-curve/backend/internal/conf"
	"io"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// LLMService 大模型服务接口
type LLMService interface {
	GenerateReviewQuestions(ctx context.Context, word, meaning, grade string) ([]*ReviewQuestion, error)
}

type llmService struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
	log     *log.Helper
}

// ReviewQuestion 复习题目
type ReviewQuestion struct {
	Type          string   `json:"type"`           // 题目类型: "multiple_choice" 或 "fill_blank"
	Question      string   `json:"question"`       // 题目内容
	Options       []string `json:"options"`        // 选择题选项（仅选择题有）
	CorrectAnswer string   `json:"correct_answer"` // 正确答案
}

// NewLLMService 创建大模型服务
func NewLLMService(lc *conf.LLM, logger log.Logger) LLMService {
	// 如果配置为 nil 或 API Key 为空，返回 nil（service 层会处理）
	if lc == nil || lc.APIKey == "" {
		log.NewHelper(logger).Warn("LLM service not configured, API Key is empty")
		return nil
	}

	baseURL := lc.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1" // 默认使用OpenAI格式
	}

	model := lc.Model
	if model == "" {
		model = "gpt-3.5-turbo" // 默认模型
	}

	return &llmService{
		apiKey:  lc.APIKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: log.NewHelper(logger),
	}
}

// OpenAIRequest OpenAI API 请求结构
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse OpenAI API 响应结构
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Error   *Error   `json:"error,omitempty"`
}

// Choice 选择结构
type Choice struct {
	Message Message `json:"message"`
}

// Error 错误结构
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// GenerateReviewQuestions 生成复习题目
func (l *llmService) GenerateReviewQuestions(ctx context.Context, word, meaning, grade string) ([]*ReviewQuestion, error) {
	if l.apiKey == "" {
		return nil, fmt.Errorf("llm api_key not configured")
	}

	// 构建提示词
	gradeHint := ""
	if grade != "" && grade != "其他" {
		gradeHint = fmt.Sprintf("（适合%s水平）", grade)
	}

	prompt := fmt.Sprintf(`请为单词 "%s"（中文释义：%s）%s生成2道复习题目，要求：
1. 第一道题：选择题（multiple_choice），包含4个选项，其中只有1个正确答案
2. 第二道题：填空题（fill_blank），要求填入该单词

%s

请以JSON格式返回，格式如下：
{
  "questions": [
    {
      "type": "multiple_choice",
      "question": "题目内容（用中文）",
      "options": ["选项A", "选项B", "选项C", "选项D"],
      "correct_answer": "正确答案"
    },
    {
      "type": "fill_blank",
      "question": "题目内容（用中文，用____表示填空位置）",
      "correct_answer": "%s"
    }
  ]
}

只返回JSON，不要其他内容。`, word, meaning, gradeHint, getGradeDifficultyHint(grade), word)

	// 构建请求
	reqBody := OpenAIRequest{
		Model: l.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	url := l.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	// 发送请求
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call llm api: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	l.log.Infof("LLM API response status: %d, body length: %d", resp.StatusCode, len(body))

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var errorResp OpenAIResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != nil {
			return nil, fmt.Errorf("llm api error: %s", errorResp.Error.Message)
		}
		return nil, fmt.Errorf("llm api returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var apiResp OpenAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("llm api error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// 提取AI返回的内容
	content := apiResp.Choices[0].Message.Content

	// 解析JSON内容
	var result struct {
		Questions []*ReviewQuestion `json:"questions"`
	}

	// 尝试提取JSON（可能包含markdown代码块）
	content = extractJSON(content)

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		l.log.Errorf("Failed to parse LLM response JSON: %v, content: %s", err, content)
		return nil, fmt.Errorf("failed to parse llm response json: %w", err)
	}

	if len(result.Questions) == 0 {
		return nil, fmt.Errorf("no questions generated")
	}

	// 确保有2道题目
	if len(result.Questions) < 2 {
		return nil, fmt.Errorf("expected 2 questions, got %d", len(result.Questions))
	}

	return result.Questions, nil
}

// extractJSON 从文本中提取JSON（处理markdown代码块等情况）
func extractJSON(text string) string {
	// 移除可能的markdown代码块标记
	text = removeMarkdownCodeBlock(text)

	// 查找JSON对象
	start := -1
	depth := 0
	for i, char := range text {
		if char == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if char == '}' {
			depth--
			if depth == 0 && start != -1 {
				return text[start : i+1]
			}
		}
	}

	// 如果没找到完整的JSON，返回原文本
	return text
}

// removeMarkdownCodeBlock 移除markdown代码块标记
func removeMarkdownCodeBlock(text string) string {
	// 移除 ```json 或 ``` 标记
	lines := []rune(text)
	result := []rune{}
	i := 0
	for i < len(lines) {
		if i+3 < len(lines) && string(lines[i:i+3]) == "```" {
			// 跳过到下一个 ```
			i += 3
			// 跳过可能的语言标识符（如 json）
			for i < len(lines) && lines[i] != '\n' {
				i++
			}
			if i < len(lines) {
				i++ // 跳过换行符
			}
			// 继续到下一个 ```
			for i < len(lines) {
				if i+3 < len(lines) && string(lines[i:i+3]) == "```" {
					i += 3
					// 跳过换行符
					if i < len(lines) && lines[i] == '\n' {
						i++
					}
					break
				}
				i++
			}
		} else {
			result = append(result, lines[i])
			i++
		}
	}
	return string(result)
}

// getGradeDifficultyHint 根据年级获取难度提示
func getGradeDifficultyHint(grade string) string {
	if grade == "" || grade == "其他" {
		return "题目难度适中，适合一般英语学习者。"
	}

	if contains(grade, "小学") {
		return "题目难度要适合小学生，使用简单的词汇和句式，题目描述要清晰易懂。"
	} else if contains(grade, "初中") {
		return "题目难度要适合初中生，使用初中水平的词汇和语法知识。"
	} else if contains(grade, "高中") {
		return "题目难度要适合高中生，可以使用较复杂的句式和词汇。"
	} else if grade == "大学" {
		return "题目难度可以较高，可以使用较复杂的词汇和句式。"
	}

	return "题目难度适中，适合一般英语学习者。"
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
