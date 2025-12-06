package conf

import (
	"github.com/go-kratos/kratos/v2/config"
)

// Bootstrap 配置结构
type Bootstrap struct {
	Server *Server `yaml:"server"`
	Data   *Data   `yaml:"data"`
}

// Server 服务器配置
type Server struct {
	HTTP *HTTP `yaml:"http"`
	GRPC *GRPC `yaml:"grpc"`
}

// HTTP HTTP 服务器配置
type HTTP struct {
	Network string        `yaml:"network"`
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
}

// GRPC gRPC 服务器配置
type GRPC struct {
	Network string        `yaml:"network"`
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
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

// Load 加载配置
func Load(c config.Config) (*Bootstrap, error) {
	var bc Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}
