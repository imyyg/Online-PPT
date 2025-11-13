# Redis 快速启动指南

本文档提供在开发环境中快速启动 Redis 的多种方式。

## 方式 1: Docker（推荐）

最简单快速的方式，无需安装 Redis。

### 启动 Redis

```bash
docker run -d \
  --name online-ppt-redis \
  -p 6379:6379 \
  redis:7-alpine
```

### 验证连接

```bash
docker exec -it online-ppt-redis redis-cli ping
```

应该返回 `PONG`。

### 停止和删除

```bash
docker stop online-ppt-redis
docker rm online-ppt-redis
```

---

## 方式 2: Docker Compose（推荐用于项目集成）

适合与项目一起管理。

### 创建 docker-compose.yml

在项目根目录创建 `docker-compose.dev.yml`:

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    container_name: online-ppt-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    restart: unless-stopped

volumes:
  redis-data:
```

### 启动服务

```bash
docker-compose -f docker-compose.dev.yml up -d
```

### 查看日志

```bash
docker-compose -f docker-compose.dev.yml logs -f redis
```

### 停止服务

```bash
docker-compose -f docker-compose.dev.yml down
```

---

## 方式 3: 本地安装（WSL/Linux）

适合需要本地持久化和系统服务管理的场景。

### Ubuntu/Debian

```bash
# 更新包索引
sudo apt update

# 安装 Redis
sudo apt install redis-server -y

# 启动 Redis
sudo systemctl start redis-server

# 设置开机自启
sudo systemctl enable redis-server

# 查看状态
sudo systemctl status redis-server
```

### 验证安装

```bash
redis-cli ping
```

应该返回 `PONG`。

### 配置文件位置

- 配置文件: `/etc/redis/redis.conf`
- 日志文件: `/var/log/redis/redis-server.log`
- 数据目录: `/var/lib/redis`

---

## 连接测试

### 使用 redis-cli

```bash
# 基本连接
redis-cli

# 测试命令
127.0.0.1:6379> ping
PONG

127.0.0.1:6379> set test "hello"
OK

127.0.0.1:6379> get test
"hello"

127.0.0.1:6379> del test
(integer) 1
```

### 使用 Go 测试连接

创建 `test-redis.go`:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

func main() {
    ctx := context.Background()
    
    // 创建 Redis 客户端
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
    
    // 测试连接
    pong, err := client.Ping(ctx).Result()
    if err != nil {
        panic(err)
    }
    fmt.Println("Redis connected:", pong)
    
    // 测试 SET/GET
    err = client.Set(ctx, "test-key", "hello from go", 10*time.Second).Err()
    if err != nil {
        panic(err)
    }
    
    val, err := client.Get(ctx, "test-key").Result()
    if err != nil {
        panic(err)
    }
    fmt.Println("Value:", val)
    
    fmt.Println("Redis test successful!")
}
```

运行测试：

```bash
go mod init test-redis
go get github.com/redis/go-redis/v9
go run test-redis.go
```

---

## 项目配置

### 更新 configs/app.yaml

```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10
```

### 环境变量（可选）

```bash
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0
```

---

## 常用 Redis 命令

### 键操作

```redis
# 查看所有键
KEYS *

# 查看特定模式的键
KEYS captcha:*
KEYS email_code:*

# 检查键是否存在
EXISTS captcha:123

# 设置键的过期时间
EXPIRE captcha:123 300

# 查看键的剩余生存时间
TTL captcha:123

# 删除键
DEL captcha:123
```

### 字符串操作

```redis
# 设置键值（带过期时间）
SET captcha:123 "AB12" EX 300

# 获取值
GET captcha:123

# 设置键值（仅当键不存在时）
SET rate_limit:user@example.com "1" EX 60 NX
```

### 调试命令

```redis
# 查看 Redis 信息
INFO

# 查看内存使用
INFO memory

# 查看客户端连接
CLIENT LIST

# 监控所有命令
MONITOR

# 清空当前数据库
FLUSHDB

# 清空所有数据库
FLUSHALL
```

---

## 故障排查

### 无法连接 Redis

1. **检查 Redis 是否运行**:
   ```bash
   # Docker
   docker ps | grep redis
   
   # 系统服务
   sudo systemctl status redis-server
   ```

2. **检查端口是否被占用**:
   ```bash
   netstat -tuln | grep 6379
   ```

3. **检查防火墙**:
   ```bash
   sudo ufw status
   sudo ufw allow 6379
   ```

### Redis 响应慢

1. **检查内存使用**:
   ```redis
   redis-cli INFO memory
   ```

2. **检查慢查询**:
   ```redis
   redis-cli SLOWLOG GET 10
   ```

3. **增加最大内存**:
   编辑 `/etc/redis/redis.conf`:
   ```conf
   maxmemory 512mb
   ```

### Docker 容器重启后数据丢失

使用持久化卷：
```bash
docker run -d \
  --name online-ppt-redis \
  -p 6379:6379 \
  -v redis-data:/data \
  redis:7-alpine redis-server --appendonly yes
```

---

## 生产环境配置建议

### 基本安全配置

编辑 `/etc/redis/redis.conf`:

```conf
# 绑定特定 IP（不对外开放）
bind 127.0.0.1

# 设置密码
requirepass your_strong_password_here

# 禁用危险命令
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG ""

# 最大内存和淘汰策略
maxmemory 256mb
maxmemory-policy allkeys-lru

# 持久化
appendonly yes
appendfsync everysec

# 日志
loglevel notice
logfile /var/log/redis/redis-server.log
```

### Docker 生产配置

```yaml
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    container_name: online-ppt-redis
    ports:
      - "127.0.0.1:6379:6379"  # 仅本地访问
    volumes:
      - redis-data:/data
      - ./redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    restart: always
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5

volumes:
  redis-data:
```

---

## 监控和维护

### 监控工具

- **Redis CLI**: `redis-cli --stat`
- **Redis Insight**: 官方 GUI 工具
- **RedisInsight**: 社区工具

### 定期维护

```bash
# 备份数据
redis-cli SAVE

# 查看数据库大小
redis-cli DBSIZE

# 查看 Redis 版本
redis-cli INFO server | grep redis_version
```

---

## 参考资源

- [Redis 官方文档](https://redis.io/docs/)
- [go-redis 文档](https://redis.uptrace.dev/)
- [Redis 命令参考](https://redis.io/commands/)
