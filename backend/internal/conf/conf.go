package conf

import (
	"time"

	"github.com/go-kratos/kratos/v2/config"
)

// Bootstrap 配置结构
type Bootstrap struct {
	Server        *Server        `yaml:"server"`
	Data          *Data          `yaml:"data"`
	Wechat        *Wechat        `yaml:"wechat"`
	OCR           *OCR           `yaml:"ocr"`           // OCR配置
	Pronunciation *Pronunciation `yaml:"pronunciation"` // 发音配置
	LLM           *LLM           `yaml:"llm"`           // 大模型配置
}

// Server 服务器配置
type Server struct {
	HTTP *HTTP `yaml:"http"`
	GRPC *GRPC `yaml:"grpc"`
}

// HTTP HTTP 服务器配置
type HTTP struct {
	Network string `yaml:"network"`
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"` // 使用字符串，例如 "5s"
}

// GetTimeout 获取超时时间
func (h *HTTP) GetTimeout() time.Duration {
	if h.Timeout == "" {
		return 5 * time.Second
	}
	d, err := time.ParseDuration(h.Timeout)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

// GRPC gRPC 服务器配置
type GRPC struct {
	Network string `yaml:"network"`
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"` // 使用字符串，例如 "5s"
}

// GetTimeout 获取超时时间
func (g *GRPC) GetTimeout() time.Duration {
	if g.Timeout == "" {
		return 5 * time.Second
	}
	d, err := time.ParseDuration(g.Timeout)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

// Data 数据配置
type Data struct {
	Database *Database `yaml:"database"`
}

// Database 数据库配置
type Database struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

// Wechat 微信配置
type Wechat struct {
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

// OCR OCR配置
type OCR struct {
	BaiduAPIKey    string `yaml:"baidu_api_key"`    // 百度 OCR API Key（可选）
	BaiduSecretKey string `yaml:"baidu_secret_key"` // 百度 OCR Secret Key（可选）
	UseTesseract   bool   `yaml:"use_tesseract"`    // 是否使用 Tesseract OCR（免费）
	TesseractPath  string `yaml:"tesseract_path"`   // Tesseract 可执行文件路径（可选，自动检测）
}

// Pronunciation 发音配置
type Pronunciation struct {
	DictionaryPath string `yaml:"dictionary_path"` // 发音字典文件路径（ultimate.json）
}

// LLM 大模型配置
type LLM struct {
	APIKey  string `yaml:"api_key"`  // 大模型API Key
	BaseURL string `yaml:"base_url"` // 大模型API Base URL（可选，默认使用OpenAI格式）
	Model   string `yaml:"model"`    // 模型名称（可选，默认gpt-3.5-turbo）
}

// Load 加载配置
func Load(c config.Config) (*Bootstrap, error) {
	var bc Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}
