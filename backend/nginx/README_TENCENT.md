# 腾讯云 SSL 证书配置说明

## 域名信息
- **域名**: `forgetting-curve.cpxdmz.top`

## 证书文件准备

### 1. 下载证书
根据 [腾讯云 SSL 证书安装文档](https://cloud.tencent.com/document/product/1207/47027)，请按以下步骤操作：

1. 登录 [SSL 证书管理控制台](https://console.cloud.tencent.com/ssl)
2. 找到对应的证书，点击"下载"
3. 选择 **Nginx** 类型的证书
4. 解压下载的证书文件

### 2. 证书文件说明
解压后会得到以下文件：
- `forgetting-curve.cpxdmz.top_bundle.crt` - 证书文件（包含证书链）
- `forgetting-curve.cpxdmz.top.key` - 私钥文件
- `forgetting-curve.cpxdmz.top.csr` - 证书签名请求文件（安装时不需要）

### 3. 放置证书文件
将以下两个文件复制到 `backend/nginx/certs/` 目录：

```bash
backend/nginx/certs/
├── forgetting-curve.cpxdmz.top_bundle.crt  # 证书文件
└── forgetting-curve.cpxdmz.top.key         # 私钥文件
```

**Windows 用户**：
```powershell
# 在项目根目录执行
copy "下载路径\forgetting-curve.cpxdmz.top_bundle.crt" "backend\nginx\certs\"
copy "下载路径\forgetting-curve.cpxdmz.top.key" "backend\nginx\certs\"
```

**Linux/macOS 用户**：
```bash
# 在项目根目录执行
cp ~/Downloads/forgetting-curve.cpxdmz.top_bundle.crt backend/nginx/certs/
cp ~/Downloads/forgetting-curve.cpxdmz.top.key backend/nginx/certs/
```

## 配置说明

### Nginx 配置
- **HTTPS 端口**: 443
- **HTTP 端口**: 80（自动跳转到 HTTPS）
- **域名**: `forgetting-curve.cpxdmz.top`
- **SSL 协议**: TLSv1, TLSv1.1, TLSv1.2, TLSv1.3
- **加密套件**: ECDHE-RSA-AES128-GCM-SHA256（参考腾讯云推荐配置）

### 自动跳转
HTTP (80端口) 请求会自动跳转到 HTTPS (443端口)，配置如下：
```nginx
server {
    listen 80;
    server_name forgetting-curve.cpxdmz.top;
    return 301 https://$host$request_uri;
}
```

## 启动服务

1. **放置证书文件后，启动服务**：
```bash
cd backend
docker-compose down
docker-compose up -d
```

2. **查看 nginx 日志**：
```bash
docker-compose logs nginx
```

3. **验证配置**：
```bash
# 检查 nginx 配置是否正确
docker exec forgetting-curve-nginx nginx -t
```

## 访问服务

配置完成后，可以通过以下方式访问：
- **HTTPS**: `https://forgetting-curve.cpxdmz.top`
- **HTTP**: `http://forgetting-curve.cpxdmz.top`（自动跳转到 HTTPS）

## 注意事项

1. **防火墙设置**：确保服务器防火墙已开放 443 和 80 端口
2. **DNS 解析**：确保域名 `forgetting-curve.cpxdmz.top` 已正确解析到服务器 IP
3. **证书有效期**：注意证书的有效期，到期前需要续期
4. **文件权限**：确保证书文件权限正确（建议 644）

## 故障排查

### 证书文件不存在
如果启动时提示证书文件不存在，检查：
```bash
ls -la backend/nginx/certs/
```

应该看到：
- `forgetting-curve.cpxdmz.top_bundle.crt`
- `forgetting-curve.cpxdmz.top.key`

### 查看 nginx 错误日志
```bash
docker exec forgetting-curve-nginx cat /var/log/nginx/error.log
```

### 测试 SSL 连接
```bash
# 测试 HTTPS 连接
curl -v https://forgetting-curve.cpxdmz.top

# 检查证书信息
openssl s_client -connect forgetting-curve.cpxdmz.top:443 -servername forgetting-curve.cpxdmz.top
```

## 参考文档
- [腾讯云 Nginx 服务器证书安装文档](https://cloud.tencent.com/document/product/1207/47027)



