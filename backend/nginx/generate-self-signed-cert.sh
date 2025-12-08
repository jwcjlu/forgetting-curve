#!/bin/bash

# 生成自签名 SSL 证书脚本
# 用于开发和测试环境

CERT_DIR="./certs"
DAYS=365

# 创建证书目录
mkdir -p "$CERT_DIR"

# 生成私钥
openssl genrsa -out "$CERT_DIR/key.pem" 2048

# 生成证书签名请求
openssl req -new -key "$CERT_DIR/key.pem" -out "$CERT_DIR/cert.csr" \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=ForgettingCurve/CN=localhost"

# 生成自签名证书
openssl x509 -req -days $DAYS -in "$CERT_DIR/cert.csr" -signkey "$CERT_DIR/key.pem" \
  -out "$CERT_DIR/cert.pem" \
  -extensions v3_req \
  -extfile <(cat <<EOF
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
)

# 清理临时文件
rm -f "$CERT_DIR/cert.csr"

echo "SSL 证书已生成到 $CERT_DIR 目录"
echo "证书文件: $CERT_DIR/cert.pem"
echo "私钥文件: $CERT_DIR/key.pem"
echo ""
echo "注意: 这是自签名证书，浏览器会显示安全警告。"
echo "生产环境请使用由 CA 签发的正式证书。"

