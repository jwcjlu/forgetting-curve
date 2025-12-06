# 构建阶段
FROM golang:1.24-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git make protobuf protobuf-dev

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 安装 proto 和 wire 工具
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest && \
    go install github.com/google/wire/cmd/wire@latest

# 设置 PATH 包含 Go bin 目录
ENV PATH=$PATH:/go/bin

# 复制源代码
COPY . .

# 生成 proto 代码
RUN cd /build/api/student/v1 && \
    protoc --proto_path=. \
    --proto_path=/build/third_party \
    --go_out=paths=source_relative:. \
    common.proto && \
    protoc --proto_path=. \
    --proto_path=/build/third_party \
    --go_out=paths=source_relative:. \
    error.proto && \
    protoc --proto_path=. \
    --proto_path=/build/third_party \
    --go_out=paths=source_relative:. \
    --go-http_out=paths=source_relative:. \
    --go-grpc_out=paths=source_relative:. \
    student.proto

# 生成 wire 代码
RUN cd /build/cmd/server && /go/bin/wire

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o server \
    ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/server .
COPY --from=builder /build/configs ./configs

# 使用 docker 配置文件覆盖默认配置
RUN if [ -f configs/config.docker.yaml ]; then \
        cp configs/config.docker.yaml configs/config.yaml && \
        echo "Using config.docker.yaml"; \
    else \
        echo "Warning: config.docker.yaml not found, using default config.yaml"; \
    fi

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# 暴露端口
EXPOSE 8000 9000

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8000 || exit 1

# 启动应用
CMD ["./server", "-conf", "configs"]

