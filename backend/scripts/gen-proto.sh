#!/bin/bash

# 生成 proto 文件
echo "Generating proto files..."

# 生成 conf proto
protoc --proto_path=./internal/conf \
  --proto_path=./third_party \
  --go_out=paths=source_relative:./internal/conf \
  ./internal/conf/*.proto

# 生成 student proto
protoc --proto_path=./api \
  --proto_path=./third_party \
  --go_out=paths=source_relative:./api \
  --go-grpc_out=paths=source_relative:./api \
  --go-http_out=paths=source_relative:./api \
  ./api/student/v1/*.proto

echo "Proto files generated successfully!"

