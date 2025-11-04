# 快速开始指南

本指南将帮助你快速开始使用 MySQL 数据库同步工具。

## 1. 准备工作

### 环境要求

- Go 1.21+ （用于编译）
- MySQL 5.7+ （源数据库和目标数据库）
- 网络连接到两个数据库

### 下载项目

```bash
git clone https://github.com/yourusername/db-sync-tools.git
cd db-sync-tools
```

## 2. 配置

### 创建配置文件

```bash
# 复制示例配置
cp config.example.yaml config.yaml
```

### 编辑配置文件

用文本编辑器打开 `config.yaml`，修改以下内容：

```yaml
# 源数据库（要复制的数据库）
source:
  url: "源数据库IP:3306/源数据库名"  # 格式: host:port/database
  username: "数据库用户名"
  password: "数据库密码"

# 目标数据库（要写入的数据库）
target:
  url: "目标数据库IP:3306/目标数据库名"  # 格式: host:port/database
  username: "数据库用户名"
  password: "数据库密码"

# 同步配置
sync:
  batch_size: 200                    # 每次传输的记录数（可根据带宽调整）
  truncate_before_sync: false        # 是否在同步前清空目标表
  exclude_tables: []                 # 要排除的表
  include_tables: []                 # 只同步这些表（空则同步所有）
  timeout: 3600                      # 超时时间（秒）
  verbose: true                      # 详细日志

# 日志配置
log:
  level: "info"                      # 日志级别: debug, info, warn, error
  file: "sync.log"                   # 日志文件
  console: true                      # 是否输出到控制台
```

## 3. 编译

### Windows

```cmd
# 使用批处理脚本
build.bat

# 或使用 go 命令
go build -o db-sync.exe cmd/db-sync/main.go
```

### Linux/Mac

```bash
# 使用 shell 脚本
chmod +x build.sh
./build.sh

# 或使用 go 命令
go build -o db-sync cmd/db-sync/main.go
```

### 使用 Makefile

```bash
make build
```

## 4. 运行

### Windows

```cmd
db-sync.exe -config config.yaml
```

### Linux/Mac

```bash
./db-sync -config config.yaml
```

## 5. 输出示例

```
=================================
MySQL Database Sync Tool v1.0.0
=================================
2025-11-04 10:30:00 [INFO] 连接源数据库: root@192.168.1.100:3306/source_db
2025-11-04 10:30:01 [INFO] 源数据库连接成功
2025-11-04 10:30:01 [INFO] 连接目标数据库: root@192.168.1.200:3306/target_db
2025-11-04 10:30:02 [INFO] 目标数据库连接成功
2025-11-04 10:30:02 [INFO] 开始发现数据库表...
2025-11-04 10:30:02 [INFO] 发现 10 个需要同步的表
2025-11-04 10:30:02 [INFO] ==============================
2025-11-04 10:30:02 [INFO] 开始同步数据库
2025-11-04 10:30:02 [INFO] 源数据库: root@192.168.1.100:3306/source_db
2025-11-04 10:30:02 [INFO] 目标数据库: root@192.168.1.200:3306/target_db
2025-11-04 10:30:02 [INFO] 批量大小: 200
2025-11-04 10:30:02 [INFO] 总表数: 10
2025-11-04 10:30:02 [INFO] ==============================
2025-11-04 10:30:02 [INFO] 总数据行数: 50000
2025-11-04 10:30:02 [INFO] 正在同步第 1/10 个表...
2025-11-04 10:30:02 [INFO] 开始同步表: users (总行数: 10000)
2025-11-04 10:30:03 [INFO] 表 users: 批次 1/50, 进度: 2.00% (200/10000), 速度: 200 行/秒
2025-11-04 10:30:04 [INFO] 表 users: 批次 2/50, 进度: 4.00% (400/10000), 速度: 200 行/秒
...
2025-11-04 10:30:52 [INFO] 表 users 同步完成: 共同步 10000 行, 耗时 50s
...
2025-11-04 10:35:00 [INFO] ==============================
2025-11-04 10:35:00 [INFO] 数据库同步完成!
2025-11-04 10:35:00 [INFO] 总表数: 10
2025-11-04 10:35:00 [INFO] 成功: 10
2025-11-04 10:35:00 [INFO] 失败: 0
2025-11-04 10:35:00 [INFO] 总行数: 50000
2025-11-04 10:35:00 [INFO] 同步行数: 50000
2025-11-04 10:35:00 [INFO] 总耗时: 4m58s
2025-11-04 10:35:00 [INFO] 平均速度: 168 行/秒
2025-11-04 10:35:00 [INFO] ==============================
2025-11-04 10:35:00 [INFO] 程序正常退出
```

## 6. 常见使用场景

### 场景1: 完全复制数据库（覆盖）

适合：初次同步或需要完全替换目标数据库

```yaml
sync:
  batch_size: 200
  truncate_before_sync: true    # 清空目标表
  exclude_tables: []
  include_tables: []             # 同步所有表
```

### 场景2: 同步特定表

适合：只需要同步部分表

```yaml
sync:
  batch_size: 200
  truncate_before_sync: true
  include_tables:
    - "users"
    - "orders"
    - "products"
```

### 场景3: 排除某些表

适合：大部分表都要同步，但排除一些表

```yaml
sync:
  batch_size: 200
  truncate_before_sync: false
  exclude_tables:
    - "tmp_*"          # 排除所有临时表
    - "cache_*"        # 排除所有缓存表
    - "log_*"          # 排除所有日志表
```

### 场景4: 低带宽环境

适合：网络带宽有限的情况

```yaml
sync:
  batch_size: 50              # 减小批量大小
  truncate_before_sync: true
  timeout: 7200               # 增加超时时间
```

### 场景5: 高速网络环境

适合：内网或高速网络

```yaml
sync:
  batch_size: 1000            # 增大批量大小
  truncate_before_sync: true
  timeout: 3600
```

## 7. 优雅退出

如果需要中断同步过程，可以使用：

- **Windows**: 按 `Ctrl+C`
- **Linux/Mac**: 按 `Ctrl+C` 或发送 SIGTERM 信号

程序会优雅退出，完成当前批次后停止。

## 8. 查看日志

日志文件默认保存在 `sync.log`：

```bash
# Linux/Mac
tail -f sync.log

# Windows (PowerShell)
Get-Content sync.log -Wait
```

## 9. 故障排查

### 问题1: 连接数据库失败

**错误信息**: `连接数据库失败: dial tcp: connect: connection refused`

**解决方法**:
- 检查数据库地址和端口是否正确
- 确认数据库服务是否运行
- 检查防火墙设置

### 问题2: 认证失败

**错误信息**: `连接数据库失败: Error 1045: Access denied`

**解决方法**:
- 确认用户名和密码是否正确
- 确认用户有足够的权限（SELECT, INSERT, CREATE, DROP）

### 问题3: 同步速度很慢

**解决方法**:
- 增加 `batch_size` 的值
- 检查网络连接速度
- 检查数据库服务器负载

### 问题4: 内存使用过高

**解决方法**:
- 减小 `batch_size` 的值
- 检查是否有大字段（BLOB, TEXT）

## 10. 最佳实践

1. **测试环境先试**: 在生产环境使用前，先在测试环境验证
2. **备份数据**: 同步前务必备份目标数据库
3. **低峰期执行**: 在业务低峰期执行同步任务
4. **逐步调优**: 从小的 batch_size 开始，逐步增加
5. **监控日志**: 实时监控日志，及时发现问题
6. **网络稳定**: 确保网络连接稳定，避免中断

## 11. 高级用法

### 自动化脚本

**Windows (PowerShell)**:

```powershell
# sync.ps1
$ErrorActionPreference = "Stop"

Write-Host "开始数据库同步..." -ForegroundColor Green

.\db-sync.exe -config config.yaml

if ($LASTEXITCODE -eq 0) {
    Write-Host "同步成功完成!" -ForegroundColor Green
} else {
    Write-Host "同步失败!" -ForegroundColor Red
    exit 1
}
```

**Linux/Mac (Bash)**:

```bash
#!/bin/bash
# sync.sh

set -e

echo "开始数据库同步..."

./db-sync -config config.yaml

if [ $? -eq 0 ]; then
    echo "同步成功完成!"
else
    echo "同步失败!"
    exit 1
fi
```

### 定时任务

**Linux (Crontab)**:

```bash
# 每天凌晨2点执行同步
0 2 * * * cd /path/to/db-sync-tools && ./db-sync -config config.yaml >> cron.log 2>&1
```

**Windows (任务计划程序)**:

创建一个批处理文件 `sync-scheduled.bat`:

```batch
@echo off
cd /d "C:\path\to\db-sync-tools"
db-sync.exe -config config.yaml >> cron.log 2>&1
```

然后在任务计划程序中添加该脚本。

## 需要帮助？

如果遇到问题，请：

1. 查看详细的 [README.md](README.md)
2. 检查日志文件 `sync.log`
3. 提交 Issue 到 GitHub

祝使用愉快！ 🎉

