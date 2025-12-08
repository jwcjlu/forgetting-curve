@echo off
REM Windows 批处理脚本：生成自签名 SSL 证书
REM 需要安装 OpenSSL for Windows

set CERT_DIR=certs
set DAYS=365

REM 创建证书目录
if not exist "%CERT_DIR%" mkdir "%CERT_DIR%"

REM 检查 OpenSSL 是否安装
where openssl >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 未找到 OpenSSL，请先安装 OpenSSL for Windows
    echo 下载地址: https://slproweb.com/products/Win32OpenSSL.html
    pause
    exit /b 1
)

REM 生成私钥
echo 正在生成私钥...
openssl genrsa -out "%CERT_DIR%\key.pem" 2048
if %ERRORLEVEL% NEQ 0 (
    echo 生成私钥失败
    pause
    exit /b 1
)

REM 生成证书签名请求
echo 正在生成证书签名请求...
openssl req -new -key "%CERT_DIR%\key.pem" -out "%CERT_DIR%\cert.csr" ^
  -subj "/C=CN/ST=Beijing/L=Beijing/O=ForgettingCurve/CN=localhost"
if %ERRORLEVEL% NEQ 0 (
    echo 生成证书签名请求失败
    pause
    exit /b 1
)

REM 创建配置文件
echo [v3_req] > "%CERT_DIR%\v3_ext.conf"
echo keyUsage = keyEncipherment, dataEncipherment >> "%CERT_DIR%\v3_ext.conf"
echo extendedKeyUsage = serverAuth >> "%CERT_DIR%\v3_ext.conf"
echo subjectAltName = @alt_names >> "%CERT_DIR%\v3_ext.conf"
echo [alt_names] >> "%CERT_DIR%\v3_ext.conf"
echo DNS.1 = localhost >> "%CERT_DIR%\v3_ext.conf"
echo DNS.2 = *.localhost >> "%CERT_DIR%\v3_ext.conf"
echo IP.1 = 127.0.0.1 >> "%CERT_DIR%\v3_ext.conf"
echo IP.2 = ::1 >> "%CERT_DIR%\v3_ext.conf"

REM 生成自签名证书
echo 正在生成自签名证书...
openssl x509 -req -days %DAYS% -in "%CERT_DIR%\cert.csr" -signkey "%CERT_DIR%\key.pem" ^
  -out "%CERT_DIR%\cert.pem" -extensions v3_req -extfile "%CERT_DIR%\v3_ext.conf"
if %ERRORLEVEL% NEQ 0 (
    echo 生成证书失败
    pause
    exit /b 1
)

REM 清理临时文件
del "%CERT_DIR%\cert.csr" "%CERT_DIR%\v3_ext.conf" 2>nul

echo.
echo SSL 证书已生成到 %CERT_DIR% 目录
echo 证书文件: %CERT_DIR%\cert.pem
echo 私钥文件: %CERT_DIR%\key.pem
echo.
echo 注意: 这是自签名证书，浏览器会显示安全警告。
echo 生产环境请使用由 CA 签发的正式证书。
echo.
pause

