package conf

import (
	"time"

	"github.com/go-kratos/kratos/v2/config"
)

// Bootstrap 配置结构
type Bootstrap struct {
	Server *Server `yaml:"server"`
	Data   *Data   `yaml:"data"`
	Wechat *Wechat `yaml:"wechat"`
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

// Load 加载配置
func Load(c config.Config) (*Bootstrap, error) {
	var bc Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}
