#!/bin/sh
set -e

CERT_DIR="/etc/nginx/certs"
# 腾讯云证书文件路径
TENCENT_CERT="$CERT_DIR/forgetting-curve.cpxdmz.top_bundle.crt"
TENCENT_KEY="$CERT_DIR/forgetting-curve.cpxdmz.top.key"
# 备用自签名证书路径
FALLBACK_CERT="$CERT_DIR/cert.pem"
FALLBACK_KEY="$CERT_DIR/key.pem"

# 确保证书目录存在
mkdir -p "$CERT_DIR"

# 检查腾讯云证书是否存在
if [ -f "$TENCENT_CERT" ] && [ -f "$TENCENT_KEY" ]; then
    echo "检测到腾讯云 SSL 证书，使用正式证书"
else
    echo "未找到腾讯云 SSL 证书"
    echo "请将以下文件放置到 nginx/certs/ 目录："
    echo "  - forgetting-curve.cpxdmz.top_bundle.crt (证书文件)"
    echo "  - forgetting-curve.cpxdmz.top.key (私钥文件)"
    echo ""
    
    # 如果备用证书也不存在，生成自签名证书（仅用于开发测试）
    if [ ! -f "$FALLBACK_CERT" ] || [ ! -f "$FALLBACK_KEY" ]; then
        echo "正在生成临时自签名证书（仅用于开发测试）..."
        
        # 生成私钥
        openssl genrsa -out "$FALLBACK_KEY" 2048
        
        # 生成证书签名请求
        openssl req -new -key "$FALLBACK_KEY" -out "$CERT_DIR/cert.csr" \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=ForgettingCurve/CN=forgetting-curve.cpxdmz.top"
        
        # 创建扩展配置文件
        cat > "$CERT_DIR/v3_ext.conf" <<EOF
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = forgetting-curve.cpxdmz.top
DNS.2 = *.forgetting-curve.cpxdmz.top
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
        
        # 生成自签名证书
        openssl x509 -req -days 365 -in "$CERT_DIR/cert.csr" -signkey "$FALLBACK_KEY" \
            -out "$FALLBACK_CERT" -extensions v3_req -extfile "$CERT_DIR/v3_ext.conf"
        
        # 清理临时文件
        rm -f "$CERT_DIR/cert.csr" "$CERT_DIR/v3_ext.conf"
        
        echo "临时自签名证书已生成（浏览器会显示安全警告）"
    fi
fi

# 执行原始的 nginx 启动命令
exec /docker-entrypoint.sh "$@"

