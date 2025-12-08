#!/bin/sh
# 不设置 set -e，以便捕获错误并输出日志

CERT_DIR="/etc/nginx/certs"
# 腾讯云证书文件路径
TENCENT_CERT="$CERT_DIR/forgetting-curve.cpxdmz.top_bundle.crt"
TENCENT_KEY="$CERT_DIR/forgetting-curve.cpxdmz.top.key"
# 备用自签名证书路径
FALLBACK_CERT="$CERT_DIR/cert.pem"
FALLBACK_KEY="$CERT_DIR/key.pem"

echo "=========================================="
echo "Nginx 容器启动脚本开始执行"
echo "=========================================="

# 确保证书目录存在
echo "[1/5] 检查证书目录..."
mkdir -p "$CERT_DIR"
if [ $? -eq 0 ]; then
    echo "✓ 证书目录已创建/存在: $CERT_DIR"
else
    echo "✗ 创建证书目录失败: $CERT_DIR"
    exit 1
fi

# 检查腾讯云证书是否存在
echo "[2/5] 检查 SSL 证书文件..."
if [ -f "$TENCENT_CERT" ] && [ -f "$TENCENT_KEY" ]; then
    echo "✓ 检测到腾讯云 SSL 证书"
    echo "  证书文件: $TENCENT_CERT"
    echo "  私钥文件: $TENCENT_KEY"
    USE_TENCENT_CERT=true
else
    echo "⚠ 未找到腾讯云 SSL 证书"
    echo "  请将以下文件放置到 nginx/certs/ 目录："
    echo "    - forgetting-curve.cpxdmz.top_bundle.crt (证书文件)"
    echo "    - forgetting-curve.cpxdmz.top.key (私钥文件)"
    USE_TENCENT_CERT=false
    
    # 如果备用证书也不存在，生成自签名证书（仅用于开发测试）
    if [ ! -f "$FALLBACK_CERT" ] || [ ! -f "$FALLBACK_KEY" ]; then
        echo "[3/5] 生成临时自签名证书（仅用于开发测试）..."
        
        # 检查 openssl 是否可用
        if ! command -v openssl >/dev/null 2>&1; then
            echo "✗ 错误: openssl 命令不可用，无法生成证书"
            echo "   请手动放置证书文件或使用包含 openssl 的镜像"
            exit 1
        fi
        
        # 生成私钥
        echo "  生成私钥..."
        if openssl genrsa -out "$FALLBACK_KEY" 2048 2>&1; then
            echo "  ✓ 私钥生成成功"
        else
            echo "  ✗ 私钥生成失败"
            exit 1
        fi
        
        # 生成证书签名请求
        echo "  生成证书签名请求..."
        if openssl req -new -key "$FALLBACK_KEY" -out "$CERT_DIR/cert.csr" \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=ForgettingCurve/CN=forgetting-curve.cpxdmz.top" 2>&1; then
            echo "  ✓ 证书签名请求生成成功"
        else
            echo "  ✗ 证书签名请求生成失败"
            exit 1
        fi
        
        # 创建扩展配置文件
        echo "  创建扩展配置..."
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
        echo "  生成自签名证书..."
        if openssl x509 -req -days 365 -in "$CERT_DIR/cert.csr" -signkey "$FALLBACK_KEY" \
            -out "$FALLBACK_CERT" -extensions v3_req -extfile "$CERT_DIR/v3_ext.conf" 2>&1; then
            echo "  ✓ 自签名证书生成成功"
        else
            echo "  ✗ 自签名证书生成失败"
            exit 1
        fi
        
        # 清理临时文件
        rm -f "$CERT_DIR/cert.csr" "$CERT_DIR/v3_ext.conf"
        echo "  ⚠ 临时自签名证书已生成（浏览器会显示安全警告）"
    else
        echo "[3/5] 使用已存在的备用证书"
    fi
fi

# 验证 nginx 配置文件
echo "[4/5] 验证 Nginx 配置文件..."

# 如果腾讯云证书不存在，创建临时证书用于配置验证
TEMP_CERT_CREATED=false
if [ ! -f "$TENCENT_CERT" ] || [ ! -f "$TENCENT_KEY" ]; then
    if [ -f "$FALLBACK_CERT" ] && [ -f "$FALLBACK_KEY" ]; then
        # 如果备用证书存在，创建符号链接或复制用于验证
        if [ ! -f "$TENCENT_CERT" ]; then
            cp "$FALLBACK_CERT" "$TENCENT_CERT" 2>/dev/null || true
            TEMP_CERT_CREATED=true
        fi
        if [ ! -f "$TENCENT_KEY" ]; then
            cp "$FALLBACK_KEY" "$TENCENT_KEY" 2>/dev/null || true
        fi
        echo "  使用备用证书进行配置验证"
    fi
fi

# 验证配置
echo "  执行 nginx -t..."
NGINX_TEST_OUTPUT=$(nginx -t 2>&1)
NGINX_TEST_EXIT_CODE=$?

# 如果创建了临时证书，清理它
if [ "$TEMP_CERT_CREATED" = true ]; then
    rm -f "$TENCENT_CERT" "$TENCENT_KEY" 2>/dev/null || true
fi

if [ $NGINX_TEST_EXIT_CODE -eq 0 ]; then
    echo "✓ Nginx 配置文件验证通过"
else
    echo "✗ Nginx 配置文件验证失败！"
    echo ""
    echo "错误详情："
    echo "$NGINX_TEST_OUTPUT"
    echo ""
    echo "常见问题排查："
    echo "  1. 证书文件不存在或路径错误"
    echo "     检查文件: $TENCENT_CERT"
    echo "     检查文件: $TENCENT_KEY"
    echo "  2. 证书文件权限不正确"
    echo "     建议权限: 644 (证书), 600 (私钥)"
    echo "  3. nginx.conf 语法错误"
    echo "     检查配置文件语法"
    echo ""
    echo "当前证书目录内容："
    ls -la "$CERT_DIR" 2>&1 || echo "  无法列出目录"
    echo ""
    exit 1
fi

# 列出证书目录内容（用于调试）
echo "[5/5] 证书目录内容："
ls -la "$CERT_DIR" 2>&1 || echo "  无法列出证书目录内容"

echo "=========================================="
echo "准备启动 Nginx..."
echo "=========================================="

# 执行原始的 nginx 启动命令
echo "启动 Nginx 服务..."
exec /docker-entrypoint.sh "$@"

