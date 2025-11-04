# 使用示例

本文档提供了各种实际使用场景的配置示例。

## 示例 1: 本地开发环境同步

**场景**: 将生产数据库同步到本地开发环境

```yaml
# config-dev-sync.yaml
source:
  host: "production-db.example.com"
  port: 3306
  username: "readonly_user"
  password: "safe_password"
  database: "production_db"
  charset: "utf8mb4"

target:
  host: "localhost"
  port: 3306
  username: "root"
  password: "local_password"
  database: "dev_db"
  charset: "utf8mb4"

sync:
  batch_size: 500
  truncate_before_sync: true
  exclude_tables:
    - "logs_*"
    - "audit_*"
    - "session_*"
  timeout: 3600
  verbose: true

log:
  level: "info"
  file: "dev-sync.log"
  console: true
```

**运行**:
```bash
./db-sync -config config-dev-sync.yaml
```

---

## 示例 2: 跨区域数据中心同步

**场景**: 从主数据中心同步到灾备数据中心

```yaml
# config-dr-sync.yaml
source:
  host: "10.1.1.100"  # 主数据中心
  port: 3306
  username: "replication_user"
  password: "strong_password"
  database: "main_db"
  charset: "utf8mb4"

target:
  host: "10.2.1.100"  # 灾备数据中心
  port: 3306
  username: "replication_user"
  password: "strong_password"
  database: "dr_db"
  charset: "utf8mb4"

sync:
  batch_size: 200  # 跨区域网络，使用较小批量
  truncate_before_sync: false
  exclude_tables: []
  include_tables: []
  timeout: 7200  # 2小时超时
  verbose: true

log:
  level: "info"
  file: "dr-sync.log"
  console: true
```

---

## 示例 3: 只同步用户相关表

**场景**: 只同步特定业务模块的表

```yaml
# config-user-tables.yaml
source:
  host: "192.168.1.100"
  port: 3306
  username: "sync_user"
  password: "password123"
  database: "app_db"
  charset: "utf8mb4"

target:
  host: "192.168.1.200"
  port: 3306
  username: "sync_user"
  password: "password123"
  database: "app_db_backup"
  charset: "utf8mb4"

sync:
  batch_size: 1000
  truncate_before_sync: true
  exclude_tables: []
  include_tables:
    - "users"
    - "user_profiles"
    - "user_settings"
    - "user_sessions"
    - "user_permissions"
  timeout: 3600
  verbose: true

log:
  level: "debug"
  file: "user-tables-sync.log"
  console: true
```

---

## 示例 4: 排除大型日志表

**场景**: 同步所有表但排除日志和临时表

```yaml
# config-exclude-logs.yaml
source:
  host: "db.example.com"
  port: 3306
  username: "admin"
  password: "admin_password"
  database: "business_db"
  charset: "utf8mb4"

target:
  host: "backup-db.example.com"
  port: 3306
  username: "admin"
  password: "admin_password"
  database: "business_db_backup"
  charset: "utf8mb4"

sync:
  batch_size: 500
  truncate_before_sync: true
  exclude_tables:
    - "access_logs"
    - "error_logs"
    - "audit_logs"
    - "tmp_*"
    - "temp_*"
    - "cache_*"
    - "_*"  # 排除所有下划线开头的表
  timeout: 3600
  verbose: false

log:
  level: "warn"
  file: "exclude-logs-sync.log"
  console: true
```

---

## 示例 5: 低带宽网络配置

**场景**: 通过VPN或低速网络同步

```yaml
# config-low-bandwidth.yaml
source:
  host: "remote-office.example.com"
  port: 3306
  username: "sync_user"
  password: "secure_pwd"
  database: "remote_db"
  charset: "utf8mb4"

target:
  host: "localhost"
  port: 3306
  username: "root"
  password: "local_pwd"
  database: "local_db"
  charset: "utf8mb4"

sync:
  batch_size: 50   # 非常小的批量
  truncate_before_sync: true
  exclude_tables: []
  include_tables: []
  timeout: 14400  # 4小时超时，给予充足时间
  verbose: true

log:
  level: "info"
  file: "low-bandwidth-sync.log"
  console: true
```

---

## 示例 6: 高速内网同步

**场景**: 在同一局域网内高速同步大量数据

```yaml
# config-high-speed.yaml
source:
  host: "192.168.10.100"
  port: 3306
  username: "admin"
  password: "password"
  database: "large_db"
  charset: "utf8mb4"

target:
  host: "192.168.10.200"
  port: 3306
  username: "admin"
  password: "password"
  database: "large_db_copy"
  charset: "utf8mb4"

sync:
  batch_size: 5000  # 大批量
  truncate_before_sync: true
  exclude_tables: []
  include_tables: []
  timeout: 3600
  verbose: true

log:
  level: "info"
  file: "high-speed-sync.log"
  console: true
```

---

## 示例 7: 增量数据追加（不清空）

**场景**: 追加新数据到目标表，不删除现有数据

```yaml
# config-incremental.yaml
source:
  host: "source-db.example.com"
  port: 3306
  username: "reader"
  password: "read_password"
  database: "source_db"
  charset: "utf8mb4"

target:
  host: "target-db.example.com"
  port: 3306
  username: "writer"
  password: "write_password"
  database: "target_db"
  charset: "utf8mb4"

sync:
  batch_size: 200
  truncate_before_sync: false  # 不清空，追加数据
  exclude_tables: []
  include_tables: []
  timeout: 3600
  verbose: true

log:
  level: "info"
  file: "incremental-sync.log"
  console: true
```

**注意**: 增量追加可能导致主键冲突，请确保数据不重复。

---

## 示例 8: 测试环境快速验证

**场景**: 小规模数据快速测试同步功能

```yaml
# config-test.yaml
source:
  host: "localhost"
  port: 3306
  username: "root"
  password: "root"
  database: "test_source"
  charset: "utf8mb4"

target:
  host: "localhost"
  port: 3306
  username: "root"
  password: "root"
  database: "test_target"
  charset: "utf8mb4"

sync:
  batch_size: 100
  truncate_before_sync: true
  exclude_tables: []
  include_tables:
    - "small_table"
  timeout: 300  # 5分钟
  verbose: true

log:
  level: "debug"
  file: "test-sync.log"
  console: true
```

---

## 示例 9: Docker 容器间同步

**场景**: 在 Docker 容器之间同步数据

```yaml
# config-docker.yaml
source:
  host: "mysql-source"  # Docker 容器名或服务名
  port: 3306
  username: "root"
  password: "mysql_root_pwd"
  database: "app_db"
  charset: "utf8mb4"

target:
  host: "mysql-target"  # Docker 容器名或服务名
  port: 3306
  username: "root"
  password: "mysql_root_pwd"
  database: "app_db"
  charset: "utf8mb4"

sync:
  batch_size: 1000
  truncate_before_sync: true
  exclude_tables: []
  include_tables: []
  timeout: 3600
  verbose: true

log:
  level: "info"
  file: "docker-sync.log"
  console: true
```

**Docker Compose 使用**:

```yaml
# docker-compose.yaml
version: '3.8'
services:
  db-sync:
    build: .
    volumes:
      - ./config-docker.yaml:/app/config.yaml
    depends_on:
      - mysql-source
      - mysql-target
    command: ./db-sync -config config.yaml
```

---

## 示例 10: 云服务器同步

**场景**: 从云服务器（如阿里云RDS）同步到本地

```yaml
# config-cloud-to-local.yaml
source:
  host: "rm-xxxxx.mysql.rds.aliyuncs.com"
  port: 3306
  username: "cloud_user"
  password: "cloud_password"
  database: "cloud_db"
  charset: "utf8mb4"

target:
  host: "localhost"
  port: 3306
  username: "root"
  password: "local_password"
  database: "local_db"
  charset: "utf8mb4"

sync:
  batch_size: 300
  truncate_before_sync: true
  exclude_tables:
    - "temp_*"
  timeout: 7200
  verbose: true

log:
  level: "info"
  file: "cloud-to-local-sync.log"
  console: true
```

---

## 批处理脚本示例

### Windows 批处理

```batch
@echo off
REM sync-all.bat - 同步多个数据库

echo ===================================
echo MySQL 数据库批量同步脚本
echo ===================================

echo.
echo [1/3] 同步用户数据库...
db-sync.exe -config config-users.yaml
if errorlevel 1 goto error

echo.
echo [2/3] 同步订单数据库...
db-sync.exe -config config-orders.yaml
if errorlevel 1 goto error

echo.
echo [3/3] 同步产品数据库...
db-sync.exe -config config-products.yaml
if errorlevel 1 goto error

echo.
echo ===================================
echo 所有数据库同步完成！
echo ===================================
goto end

:error
echo.
echo ===================================
echo 同步失败！请检查日志。
echo ===================================
exit /b 1

:end
```

### Linux/Mac Shell 脚本

```bash
#!/bin/bash
# sync-all.sh - 同步多个数据库

set -e

echo "==================================="
echo "MySQL 数据库批量同步脚本"
echo "==================================="

databases=("users" "orders" "products")
total=${#databases[@]}

for i in "${!databases[@]}"; do
    db=${databases[$i]}
    index=$((i + 1))
    
    echo ""
    echo "[$index/$total] 同步 $db 数据库..."
    
    if ./db-sync -config "config-${db}.yaml"; then
        echo "✓ $db 数据库同步成功"
    else
        echo "✗ $db 数据库同步失败"
        exit 1
    fi
done

echo ""
echo "==================================="
echo "所有数据库同步完成！"
echo "==================================="
```

---

## 监控脚本示例

### 实时监控同步进度

```bash
#!/bin/bash
# monitor-sync.sh

LOG_FILE="sync.log"

echo "监控同步进度..."
echo "按 Ctrl+C 退出监控"
echo ""

tail -f "$LOG_FILE" | while read line; do
    # 高亮显示重要信息
    if [[ $line == *"ERROR"* ]]; then
        echo -e "\e[31m$line\e[0m"  # 红色
    elif [[ $line == *"WARN"* ]]; then
        echo -e "\e[33m$line\e[0m"  # 黄色
    elif [[ $line == *"成功"* ]] || [[ $line == *"完成"* ]]; then
        echo -e "\e[32m$line\e[0m"  # 绿色
    else
        echo "$line"
    fi
done
```

---

## 定时任务示例

### Linux Crontab

```bash
# 每天凌晨 2:00 同步
0 2 * * * cd /opt/db-sync-tools && ./db-sync -config config.yaml >> /var/log/db-sync-cron.log 2>&1

# 每小时同步一次（适合增量同步）
0 * * * * cd /opt/db-sync-tools && ./db-sync -config config-incremental.yaml >> /var/log/db-sync-hourly.log 2>&1

# 每周日凌晨 3:00 全量同步
0 3 * * 0 cd /opt/db-sync-tools && ./db-sync -config config-full.yaml >> /var/log/db-sync-weekly.log 2>&1
```

### Windows 任务计划程序（XML格式）

```xml
<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Triggers>
    <CalendarTrigger>
      <StartBoundary>2025-01-01T02:00:00</StartBoundary>
      <Enabled>true</Enabled>
      <ScheduleByDay>
        <DaysInterval>1</DaysInterval>
      </ScheduleByDay>
    </CalendarTrigger>
  </Triggers>
  <Actions>
    <Exec>
      <Command>C:\db-sync-tools\db-sync.exe</Command>
      <Arguments>-config config.yaml</Arguments>
      <WorkingDirectory>C:\db-sync-tools</WorkingDirectory>
    </Exec>
  </Actions>
</Task>
```

---

## 常见问题解决

### 问题: 特定表同步失败

创建只同步失败表的配置：

```yaml
# config-retry.yaml
sync:
  include_tables:
    - "failed_table_1"
    - "failed_table_2"
  truncate_before_sync: true
  batch_size: 100  # 使用更小的批量
```

### 问题: 需要跳过某些列

目前工具同步整表，如需跳过某些列，可以：
1. 在目标数据库创建视图
2. 或在源数据库创建只包含需要列的表

---

希望这些示例能帮助你更好地使用数据库同步工具！

