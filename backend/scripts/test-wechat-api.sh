#!/bin/bash
# 测试微信 API 配置
# 使用方法: ./test-wechat-api.sh <appid> <secret> <code>

APPID=$1
SECRET=$2
CODE=$3

if [ -z "$APPID" ] || [ -z "$SECRET" ] || [ -z "$CODE" ]; then
    echo "使用方法: $0 <appid> <secret> <code>"
    echo "示例: $0 wx2c2b1fd1b7d586c0 f3e9987492fcef649c13cc3a1cdecbb3 0e1him100GDGsV1Zmf100bWlu93him1u"
    exit 1
fi

URL="https://api.weixin.qq.com/sns/jscode2session?appid=${APPID}&secret=${SECRET}&js_code=${CODE}&grant_type=authorization_code"

echo "测试微信 API..."
echo "AppID: $APPID"
echo "Secret: ${SECRET:0:10}... (隐藏)"
echo "Code: ${CODE:0:10}... (隐藏)"
echo ""
echo "请求 URL: $URL"
echo ""

RESPONSE=$(curl -s "$URL")
echo "响应: $RESPONSE"
echo ""

# 解析响应
ERRCODE=$(echo $RESPONSE | grep -o '"errcode":[0-9]*' | grep -o '[0-9]*')
ERRMSG=$(echo $RESPONSE | grep -o '"errmsg":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ERRCODE" ] || [ "$ERRCODE" = "0" ]; then
    echo "✓ 成功！"
    OPENID=$(echo $RESPONSE | grep -o '"openid":"[^"]*"' | cut -d'"' -f4)
    echo "OpenID: $OPENID"
else
    echo "✗ 失败！"
    echo "错误码: $ERRCODE"
    echo "错误信息: $ERRMSG"
    echo ""
    case $ERRCODE in
        40029)
            echo "原因: code 无效或已过期"
            echo "建议: 1) 检查 AppID/Secret 是否正确 2) code 是否过期 3) code 是否被重复使用"
            ;;
        40013)
            echo "原因: AppID 无效"
            echo "建议: 检查 AppID 是否正确"
            ;;
        40125)
            echo "原因: AppSecret 无效"
            echo "建议: 检查 AppSecret 是否正确"
            ;;
        *)
            echo "请查看微信 API 文档了解错误详情"
            ;;
    esac
fi



