# 艾宾浩斯单词背诵系统 - 后端服务

基于 Kratos 框架开发的单词背诵系统后端服务。

## 功能特性

1. **学生管理**
   - 添加学生
   - 获取学生信息
   - 学号唯一性验证

2. **单词管理**
   - 批量添加单词到学生名下
   - 获取学生的单词列表（分页）
   - 权限控制：每个学生只能查看自己的单词

3. **数据安全**
   - 学生数据隔离
   - 自动数据验证
   - 软删除支持

## 技术栈

- **框架**: Kratos v2
- **数据库**: MySQL
- **ORM**: GORM
- **协议**: gRPC + HTTP
- **依赖注入**: Wire

## 项目结构

```
backend/
├── api/                    # API 定义
│   └── student/v1/        # 学生服务 API
├── cmd/                    # 应用入口
│   └── server/            # 服务器启动
├── configs/               # 配置文件
├── internal/              # 内部代码
│   ├── biz/              # 业务逻辑层
│   ├── data/             # 数据访问层
│   ├── service/          # 服务层
│   └── conf/             # 配置定义
└── go.mod                # Go 模块定义
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 5.7+
- Make (可选)

### 2. 安装依赖

```bash
# 安装 Go 依赖
go mod download

# 安装工具
make init
# 或手动安装
go install github.com/google/wire/cmd/wire@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 3. 配置数据库

修改 `configs/config.yaml` 中的数据库连接信息：

```yaml
data:
  database:
    driver: mysql
    source: root:password@tcp(localhost:3306)/forgetting_curve?charset=utf8mb4&parseTime=True&loc=Local
```

创建数据库：

```sql
CREATE DATABASE forgetting_curve CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. 生成代码

```bash
# 生成 proto 代码
make api

# 生成 wire 依赖注入代码
make wire
```

### 5. 运行服务

```bash
# 开发模式运行
make run

# 或直接运行
go run ./cmd/server -conf ../../configs
```

服务启动后：
- HTTP API: http://localhost:8000
- gRPC: localhost:9000

## API 接口

### 1. 添加学生

```http
POST /api/v1/students
Content-Type: application/json

{
  "name": "张三",
  "student_no": "2023001"
}
```

**响应：**
```json
{
  "student": {
    "id": 1,
    "name": "张三",
    "student_no": "2023001",
    "created_at": 1699000000,
    "updated_at": 1699000000
  }
}
```

### 2. 获取学生信息

```http
GET /api/v1/students/{id}
```

### 3. 批量添加单词

```http
POST /api/v1/students/{student_id}/words/batch
Content-Type: application/json

{
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
}
```

**响应：**
```json
{
  "count": 2,
  "words": [
    {
      "id": 1,
      "student_id": 1,
      "word": "hello",
      "meaning": "你好",
      "start_date": "2024-01-01",
      "review_count": 0,
      "last_review_date": "",
      "created_at": 1699000000,
      "updated_at": 1699000000
    }
  ]
}
```

### 4. 获取学生的单词列表

```http
GET /api/v1/students/{student_id}/words?page=1&page_size=20
```

**响应：**
```json
{
  "words": [...],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

## 权限控制

系统实现了严格的数据隔离：

1. **获取单词列表**：只能通过 `student_id` 参数获取，系统会验证该学生是否存在，并只返回该学生的单词
2. **批量添加单词**：必须指定 `student_id`，单词会自动关联到该学生
3. **数据验证**：所有操作都会验证学生ID的有效性

## 数据库模型

### Student (学生表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| name | varchar(100) | 学生姓名 |
| student_no | varchar(50) | 学号（唯一） |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 删除时间（软删除） |

### Word (单词表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| student_id | bigint | 学生ID（外键） |
| word | varchar(100) | 单词 |
| meaning | varchar(500) | 释义 |
| start_date | varchar(20) | 开始日期 |
| review_count | int | 复习次数 |
| last_review_date | varchar(20) | 上次复习日期 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 删除时间（软删除） |

## 开发说明

### 添加新的 API

1. 在 `api/student/v1/student.proto` 中定义新的 RPC 方法
2. 运行 `make api` 生成代码
3. 在 `internal/biz` 中实现业务逻辑
4. 在 `internal/service` 中实现服务接口
5. 在 `internal/data` 中添加数据访问方法

### 测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test ./internal/biz/...
```

## 部署

### 构建

```bash
make build
```

### Docker (可选)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server", "-conf", "configs"]
```

## 许可证

MIT License



