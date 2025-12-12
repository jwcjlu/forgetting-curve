# Docker 部署指南

## 快速开始

### 1. 使用 Docker Compose（推荐）

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

### 2. 开发环境

```bash
# 只启动数据库（后端在本地运行）
docker-compose -f docker-compose.dev.yml up -d

# 修改本地 configs/config.yaml 中的数据库连接为：
# source: appuser:apppassword@tcp(localhost:3307)/forgetting_curve?charset=utf8mb4&parseTime=True&loc=Local
```

## 服务说明

### MySQL 数据库
- **端口**: 3306 (生产) / 3307 (开发)
- **用户名**: appuser
- **密码**: apppassword
- **数据库**: forgetting_curve
- **Root 密码**: rootpassword

### 后端服务
- **HTTP 端口**: 8000
- **gRPC 端口**: 9000
- **健康检查**: http://localhost:8000/health

## 环境变量

可以通过环境变量覆盖配置：

```bash
# 在 docker-compose.yml 中设置
environment:
  - DB_HOST=mysql
  - DB_PASSWORD=your_password
```

## 数据持久化

数据存储在 Docker volume 中：
- `mysql_data`: MySQL 数据文件

查看 volumes：
```bash
docker volume ls
```

备份数据：
```bash
docker exec forgetting-curve-mysql mysqldump -u root -prootpassword forgetting_curve > backup.sql
```

恢复数据：
```bash
docker exec -i forgetting-curve-mysql mysql -u root -prootpassword forgetting_curve < backup.sql
```

## 构建自定义镜像

```bash
# 构建镜像
docker-compose build

# 或单独构建
docker build -t forgetting-curve-backend:latest .
```

## 生产环境部署

### 1. 修改配置

编辑 `docker-compose.yml`：
- 修改数据库密码
- 修改端口映射（如果需要）
- 添加环境变量

### 2. 使用环境变量文件

创建 `.env` 文件：
```bash
cp .env.example .env
# 编辑 .env 文件
```

在 `docker-compose.yml` 中使用：
```yaml
env_file:
  - .env
```

### 3. 使用外部数据库

如果使用外部数据库，可以只启动后端服务：

```yaml
services:
  backend:
    # ... 配置
    environment:
      - DB_HOST=your-db-host
      - DB_PORT=3306
      # ...
```

## 故障排查

### 查看日志
```bash
# 所有服务日志
docker-compose logs

# 特定服务日志
docker-compose logs backend
docker-compose logs mysql

# 实时日志
docker-compose logs -f backend
```

### 进入容器
```bash
# 进入后端容器
docker exec -it forgetting-curve-backend sh

# 进入数据库容器
docker exec -it forgetting-curve-mysql mysql -u root -prootpassword
```

### 检查服务状态
```bash
# 查看运行状态
docker-compose ps

# 查看健康检查
docker inspect forgetting-curve-backend | grep Health -A 10
```

## 性能优化

### 数据库连接池

在应用代码中已配置连接池，如需调整，修改 `internal/data/data.go`：

```go
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### MySQL 配置优化

创建 `my.cnf` 并挂载到容器：

```yaml
volumes:
  - ./mysql/my.cnf:/etc/mysql/conf.d/custom.cnf:ro
```

## 安全建议

1. **修改默认密码**：生产环境必须修改所有默认密码
2. **限制端口暴露**：只暴露必要的端口
3. **使用 secrets**：敏感信息使用 Docker secrets
4. **网络隔离**：使用自定义网络
5. **定期备份**：设置自动备份策略

## 扩展部署

### 添加 Redis（可选）

```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: forgetting-curve-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - app-network

volumes:
  redis_data:
```

### 添加 Nginx 反向代理

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - backend
```



