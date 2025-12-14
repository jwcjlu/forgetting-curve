# Tesseract OCR 免费方案配置指南

## 简介

Tesseract OCR 是一个**完全免费、开源**的 OCR 引擎，由 Google 开发。无需任何 API 密钥或付费服务。

## 安装 Tesseract OCR

### Windows

1. **下载安装包**
   - 访问：https://github.com/UB-Mannheim/tesseract/wiki
   - 下载最新版本的 Windows 安装包（.exe 文件）

2. **安装**
   - 运行安装程序
   - 安装路径建议：`C:\Program Files\Tesseract-OCR`
   - **重要**：安装时勾选"添加到 PATH 环境变量"

3. **验证安装**
   ```powershell
   tesseract --version
   ```

### Linux (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install tesseract-ocr
sudo apt-get install tesseract-ocr-eng  # 英文语言包
```

### macOS

```bash
brew install tesseract
```

## 配置后端

在 `backend/configs/config.yaml` 中添加 OCR 配置：

```yaml
ocr:
  use_tesseract: true  # 启用 Tesseract OCR（免费）
  tesseract_path: ""   # 留空自动检测，或指定路径如 "C:\\Program Files\\Tesseract-OCR\\tesseract.exe"
  # baidu_api_key: ""   # 不需要配置百度 OCR
  # baidu_secret_key: ""
```

## 使用说明

### 自动检测 Tesseract

如果 `tesseract_path` 为空，系统会自动尝试以下路径：
- Windows: `C:\Program Files\Tesseract-OCR\tesseract.exe`
- Linux: `/usr/bin/tesseract` 或 `/usr/local/bin/tesseract`
- 系统 PATH 中的 `tesseract` 命令

### 手动指定路径

如果自动检测失败，可以手动指定：

```yaml
ocr:
  use_tesseract: true
  tesseract_path: "C:\\Program Files\\Tesseract-OCR\\tesseract.exe"
```

## 优势

✅ **完全免费** - 无需 API 密钥，无使用限制  
✅ **开源** - 代码完全开放，可自定义  
✅ **离线工作** - 不需要网络连接  
✅ **隐私保护** - 图片不会上传到第三方服务器  
✅ **多语言支持** - 支持 100+ 种语言

## 注意事项

1. **识别精度**：Tesseract 的识别精度可能略低于商业 OCR API，但对于清晰的英文文本通常足够使用
2. **性能**：本地 OCR 处理速度取决于 CPU 性能
3. **语言包**：默认只安装英文，如需识别其他语言需要安装对应的语言包

## 故障排查

### 错误：Tesseract OCR 未安装或未找到

**解决方案**：
1. 确认 Tesseract 已正确安装
2. 检查是否添加到 PATH 环境变量
3. 手动指定 `tesseract_path` 配置

### 错误：无法识别中文

**解决方案**：
1. 安装中文语言包：
   - Windows: 重新运行安装程序，选择中文语言包
   - Linux: `sudo apt-get install tesseract-ocr-chi-sim`
2. 修改代码中的语言参数（如果需要）

## 测试

运行测试程序：

```bash
cd backend/internal/pkg/ocr
go run test_ocr.go test.jpg
```

## 与百度 OCR 对比

| 特性 | Tesseract OCR | 百度 OCR API |
|------|--------------|--------------|
| 费用 | 完全免费 | 有免费额度，超出收费 |
| 网络 | 不需要 | 需要 |
| 隐私 | 本地处理 | 上传到服务器 |
| 精度 | 良好 | 优秀 |
| 速度 | 取决于 CPU | 取决于网络 |

## 推荐

对于单词识别场景，**Tesseract OCR 完全够用**，推荐使用免费方案！




