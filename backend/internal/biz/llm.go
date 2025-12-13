package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"forgetting-curve/backend/internal/conf"
	"io"
	"net/http"
	"regexp"
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

	prompt := fmt.Sprintf(`Please generate 2 review questions for the word "%s" (meaning: %s) %s. Requirements:
1. First question: Multiple choice (multiple_choice) with 4 options, only 1 correct answer
   - Options can be other words related to this word (such as synonyms, antonyms, related words, etc.)
   - The correct answer is not necessarily "%s", it can be other related words
   - The question should test understanding and application of this word
   - ALL TEXT MUST BE IN ENGLISH (question text, options, everything)
2. Second question: Fill in the blank (fill_blank), requiring to fill in the word "%s"
   - The question should test spelling and memory of this word
   - ALL TEXT MUST BE IN ENGLISH

%s

Please return strictly in the following JSON format, no other text, only JSON:
{
  "questions": [
    {
      "type": "multiple_choice",
      "question": "Question text in English (test understanding of the word)",
      "options": ["Option A", "Option B", "Option C", "Option D"],
      "correct_answer": "Correct answer (can be "%s" or other related words)"
    },
    {
      "type": "fill_blank",
      "question": "Question text in English (use ____ to indicate blank, test spelling)",
      "correct_answer": "%s"
    }
  ]
}

IMPORTANT: 
- All question text and options MUST be in English
- Only return JSON object, do not use markdown code blocks, do not add any explanatory text
- Do not include any Chinese characters in the question text or options`, word, meaning, gradeHint, word, word, getGradeDifficultyHint(grade), word, word)

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
	req, err := http.NewRequestWithContext(context.WithoutCancel(ctx), "POST", url, bytes.NewBuffer(jsonData))
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

	// 记录原始内容用于调试
	l.log.Infof("LLM raw response content (first 500 chars): %s", truncateString(content, 500))

	// 检查内容是否为空
	if content == "" {
		return nil, fmt.Errorf("llm returned empty content")
	}

	// 解析JSON内容
	var result struct {
		Questions []*ReviewQuestion `json:"questions"`
	}

	// 尝试提取JSON（可能包含markdown代码块）
	extractedJSON := extractJSON(content)

	// 记录提取后的JSON用于调试
	l.log.Infof("Extracted JSON (first 500 chars): %s", truncateString(extractedJSON, 500))

	// 检查提取后的JSON是否为空
	if extractedJSON == "" {
		l.log.Errorf("Failed to extract JSON from content. Original content: %s", truncateString(content, 1000))
		return nil, fmt.Errorf("failed to extract json from llm response, content may not contain valid json")
	}

	if err := json.Unmarshal([]byte(extractedJSON), &result); err != nil {
		// 尝试修复常见的JSON问题
		fixedJSON := tryFixJSON(extractedJSON)
		if fixedJSON != extractedJSON {
			l.log.Infof("Attempting to fix JSON, trying: %s", truncateString(fixedJSON, 500))
			if err2 := json.Unmarshal([]byte(fixedJSON), &result); err2 == nil {
				l.log.Infof("Successfully fixed and parsed JSON")
			} else {
				l.log.Errorf("Failed to parse LLM response JSON (original error: %v, fixed error: %v), extracted JSON: %s, original content: %s",
					err, err2, truncateString(extractedJSON, 1000), truncateString(content, 1000))
				return nil, fmt.Errorf("failed to parse llm response json: %w (also tried fixing: %v)", err, err2)
			}
		} else {
			l.log.Errorf("Failed to parse LLM response JSON: %v, extracted JSON: %s, original content: %s",
				err, truncateString(extractedJSON, 1000), truncateString(content, 1000))
			return nil, fmt.Errorf("failed to parse llm response json: %w", err)
		}
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
	if text == "" {
		return ""
	}

	// 移除可能的markdown代码块标记
	text = removeMarkdownCodeBlock(text)

	// 移除前后空白字符
	text = trimSpace(text)

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
				extracted := text[start : i+1]
				// 验证提取的内容是否是有效的JSON
				if isValidJSON(extracted) {
					return extracted
				}
				// 如果不是有效JSON，继续查找下一个
				start = -1
				depth = 0
			}
		}
	}

	// 如果没找到完整的JSON，尝试返回原文本（可能是纯JSON）
	if isValidJSON(text) {
		return text
	}

	// 如果原文本也不是有效JSON，返回空字符串
	return ""
}

// trimSpace 移除字符串前后的空白字符
func trimSpace(s string) string {
	// 移除前导空白
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// 移除尾随空白
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// isValidJSON 检查字符串是否是有效的JSON
func isValidJSON(s string) bool {
	if s == "" {
		return false
	}
	var temp interface{}
	return json.Unmarshal([]byte(s), &temp) == nil
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// tryFixJSON 尝试修复常见的JSON问题
func tryFixJSON(jsonStr string) string {
	// 移除可能的尾随逗号
	jsonStr = removeTrailingCommas(jsonStr)

	// 尝试修复未转义的换行符
	jsonStr = fixUnescapedNewlines(jsonStr)

	return jsonStr
}

// removeTrailingCommas 移除JSON中的尾随逗号
func removeTrailingCommas(jsonStr string) string {
	// 使用正则表达式移除对象和数组中的尾随逗号
	// 例如: {"key": "value",} -> {"key": "value"}
	// 例如: ["item1", "item2",] -> ["item1", "item2"]
	re := regexp.MustCompile(`,\s*([}\]])`)
	return re.ReplaceAllString(jsonStr, "$1")
}

// fixUnescapedNewlines 修复未转义的换行符（简化实现）
func fixUnescapedNewlines(jsonStr string) string {
	// 这是一个简化的实现
	// 实际上，修复未转义的换行符比较复杂，需要正确解析JSON字符串
	// 这里我们只做基本的清理，移除可能导致问题的字符
	// 更复杂的修复应该由LLM本身返回正确的JSON
	return jsonStr
}

// removeMarkdownCodeBlock 移除markdown代码块标记
func removeMarkdownCodeBlock(text string) string {
	// 移除 ```json 或 ``` 标记
	lines := []rune(text)
	result := []rune{}
	i := 0
	for i < len(lines) {
		if i+2 < len(lines) && string(lines[i:i+3]) == "```" {
			// 跳过到下一个 ```
			i += 3
			// 跳过可能的语言标识符（如 json）
			for i < len(lines) && lines[i] != '\n' {
				i++
			}
			if i < len(lines) {
				i++ // 跳过换行符
			}
			// 继续到下一个 ```，保留中间的内容
			for i < len(lines) {
				if i+2 < len(lines) && string(lines[i:i+3]) == "```" {
					i += 3
					// 跳过换行符
					if i < len(lines) && lines[i] == '\n' {
						i++
					}
					break
				}
				// 保留代码块内的内容
				result = append(result, lines[i])
				i++
			}
		} else {
			result = append(result, lines[i])
			i++
		}
	}
	return string(result)
}

// getGradeDifficultyHint 根据年级获取难度提示（英文）
func getGradeDifficultyHint(grade string) string {
	if grade == "" || grade == "其他" {
		return "Question difficulty should be moderate, suitable for general English learners."
	}

	if contains(grade, "小学") {
		return "Question difficulty should be suitable for elementary school students, use simple vocabulary and sentence structures, questions should be clear and easy to understand."
	} else if contains(grade, "初中") {
		return "Question difficulty should be suitable for middle school students, use middle school level vocabulary and grammar knowledge."
	} else if contains(grade, "高中") {
		return "Question difficulty should be suitable for high school students, can use more complex sentence structures and vocabulary."
	} else if grade == "大学" {
		return "Question difficulty can be higher, can use more complex vocabulary and sentence structures."
	}

	return "Question difficulty should be moderate, suitable for general English learners."
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
