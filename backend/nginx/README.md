# Nginx 配置说明

## 概述

本目录包含 Nginx 反向代理配置，用于将 HTTPS (443端口) 请求转发到后端服务 (8000端口)。

## 文件结构

```
nginx/
├── nginx.conf              # Nginx 主配置文件
├── certs/                  # SSL 证书目录（需要手动创建）
│   ├── cert.pem           # SSL 证书文件
│   └── key.pem            # SSL 私钥文件
├── generate-self-signed-cert.sh  # 生成自签名证书脚本
└── README.md              # 本文件
```

## SSL 证书配置

### 方式一：使用自签名证书（开发/测试环境）

1. 运行生成脚本（Linux/macOS）：
```bash
cd backend/nginx
chmod +x generate-self-signed-cert.sh
./generate-self-signed-cert.sh
```

2. Windows 用户可以使用 Git Bash 或 WSL，或者手动生成：
```bash
# 创建证书目录
mkdir certs

# 生成私钥
openssl genrsa -out certs/key.pem 2048

# 生成证书
openssl req -new -x509 -key certs/key.pem -out certs/cert.pem -days 365 \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=ForgettingCurve/CN=localhost"
```

**注意**: 自签名证书会导致浏览器显示安全警告，仅适用于开发和测试环境。

### 方式二：使用正式证书（生产环境）

1. 从 CA（如 Let's Encrypt）获取证书
2. 将证书文件重命名为 `cert.pem`，私钥文件重命名为 `key.pem`
3. 放置在 `nginx/certs/` 目录下

## 配置说明

- **443 端口**: HTTPS 服务，反向代理到 `backend:8000`
- **80 端口**: HTTP 服务，自动重定向到 HTTPS
- **WebSocket 支持**: 已配置，支持 WebSocket 连接
- **Gzip 压缩**: 已启用，提升传输效率

## 使用方式

1. 确保 SSL 证书已放置在 `nginx/certs/` 目录
2. 启动服务：
```bash
docker-compose up -d
```

3. 访问服务：
   - HTTPS: `https://localhost`
   - HTTP: `http://localhost` (自动重定向到 HTTPS)

## 故障排查

### 证书文件不存在
如果 nginx 启动失败，检查证书文件是否存在：
```bash
ls -la backend/nginx/certs/
```

### 端口被占用
检查 443 和 80 端口是否被占用：
```bash
# Linux/macOS
netstat -tuln | grep -E ':(80|443)'

# Windows
netstat -ano | findstr ":443"
netstat -ano | findstr ":80"
```

### 查看 nginx 日志
```bash
docker logs forgetting-curve-nginx
```




