# 快速开始指南

## 前置要求

1. **Go 1.21+**: 安装 Go 语言环境
2. **MySQL 5.7+**: 安装并启动 MySQL 数据库
3. **protoc**: Protocol Buffers 编译器（可选，用于生成代码）

## 安装步骤

### 1. 安装依赖工具

```bash
# 安装 Wire（依赖注入）
go install github.com/google/wire/cmd/wire@latest

# 安装 protoc 插件（如果还没有安装）
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
```

### 2. 配置数据库

1. 创建数据库：

```sql
CREATE DATABASE forgetting_curve CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 修改配置文件 `configs/config.yaml`：

```yaml
data:
  database:
    driver: mysql
    source: root:your_password@tcp(localhost:3306)/forgetting_curve?charset=utf8mb4&parseTime=True&loc=Local
```

将 `your_password` 替换为你的 MySQL 密码。

### 3. 生成代码

```bash
# 进入项目目录
cd backend

# 生成 proto 代码（如果 proto 文件已修改）
# 注意：如果使用预生成的代码，可以跳过此步骤
make api

# 生成 Wire 依赖注入代码
make wire
```

### 4. 运行服务

```bash
# 开发模式运行
make run

# 或直接运行
go run ./cmd/server -conf ../../configs
```

看到以下输出表示启动成功：

```
INFO msg=server listening address=[::]:8000
INFO msg=server listening address=[::]:9000
```

## 测试 API

### 1. 添加学生

```bash
curl -X POST http://localhost:8000/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "student_no": "2023001"
  }'
```

### 2. 批量添加单词

```bash
curl -X POST http://localhost:8000/api/v1/students/1/words/batch \
  -H "Content-Type: application/json" \
  -d '{
    "words": [
      {
        "word": "hello",
        "meaning": "你好",
        "start_date": "2024-01-01"
      },
      {
        "word": "world",
        "meaning": "世界",
        "start_date": "2024-01-01"
      }
    ]
  }'
```

### 3. 获取学生的单词列表

```bash
curl http://localhost:8000/api/v1/students/1/words?page=1&page_size=20
```

## 常见问题

### Q: 提示 "failed to open database"

**A**: 检查：
1. MySQL 服务是否启动
2. 数据库是否已创建
3. 配置文件中的连接信息是否正确

### Q: 提示 "wire: no provider found"

**A**: 运行 `make wire` 生成依赖注入代码

### Q: 提示 "protoc: command not found"

**A**: 
- 如果只是运行服务，不需要 protoc（使用预生成的代码）
- 如果需要修改 proto 文件，请安装 protoc：
  - macOS: `brew install protobuf`
  - Ubuntu: `apt-get install protobuf-compiler`
  - Windows: 从 [protobuf releases](https://github.com/protocolbuffers/protobuf/releases) 下载

### Q: 如何查看 API 文档？

**A**: 可以使用 Postman 或 curl 测试 API，也可以使用 Swagger（需要额外配置）

## 下一步

- 查看 `README.md` 了解完整的 API 文档
- 修改业务逻辑以满足你的需求
- 添加认证和授权机制
- 部署到生产环境




