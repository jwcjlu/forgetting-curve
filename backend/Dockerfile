# 构建阶段
FROM golang:1.24-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git make protobuf protobuf-dev

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 生成 proto 代码（如果需要）
# RUN make api

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

# 如果存在 docker 配置文件，使用它
RUN if [ -f configs/config.docker.yaml ]; then \
        cp configs/config.docker.yaml configs/config.yaml; \
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

