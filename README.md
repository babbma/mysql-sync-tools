# MySQL 数据库同步工具

## 🤔 你是否遇到过这些问题？

作为开发者或运维人员，你可能经常遇到这些头疼的场景：

- 📦 **搭建测试环境**时，需要从生产数据库复制数据，但手动导出导入太繁琐？
- 🔄 **数据库迁移**时，需要在不同服务器间同步大量数据，担心操作失误？
- 🌐 **跨区域同步**时，网络带宽有限，一次性传输太大容易失败或超时？
- 🏢 **分支机构同步**时，需要定期将总部数据同步到各地分支，人工操作耗时耗力？
- 🔧 **没有主键的表**导致现有工具无法处理，只能写脚本手动处理？
- ⚡ **同步过程漫长**，想中途停止却发现程序卡死，只能强制结束？

传统的 `mysqldump` 虽然可靠，但面对大数据量和复杂场景时：
- ❌ 一次性导出大文件，容易因网络问题失败
- ❌ 无法实时查看进度，不知道还要等多久
- ❌ 没有断点续传，中断后需要重新开始
- ❌ 配置复杂，每次都要输入一长串命令参数

## ✨ 我们的解决方案

这个工具就是为了解决上述痛点而生：

### 核心优势

🎯 **智能批量传输**
- 自动分批传输数据（默认 200 条/批次）
- 完美应对低带宽环境，传输过程更稳定
- 可根据实际情况灵活调整批量大小

📊 **实时进度监控**
- 清晰显示：当前进度、传输速度、预计剩余时间
- 不再盲等，随时掌握同步状态
- 每个表的同步情况一目了然

⚙️ **简单易用**
- 仅需 3 个参数配置（URL、用户名、密码）
- 一次配置，重复使用
- 支持表过滤，只同步需要的数据

🛡️ **稳定可靠**
- 支持 MySQL 5.5/5.6/5.7/8.0 全版本
- 支持无主键的表（自动适配）
- Ctrl+C 随时中断，优雅退出
- 详细的错误提示和日志记录

🚀 **高性能**
- Go 语言编写，性能出色
- 批量 INSERT 优化，同步速度快
- 支持大表同步，内存占用可控

### 它能做什么？

✅ **数据库克隆**：快速克隆整个数据库到另一台服务器  
✅ **环境搭建**：一键从生产环境同步数据到测试/开发环境  
✅ **数据备份**：定期将数据同步到备份服务器  
✅ **灾备同步**：跨地域数据中心的数据同步  
✅ **分支同步**：总部数据定期同步到各分支机构  

### 真实场景示例

**场景 1：测试环境搭建**
```
问题：需要从生产环境复制 50 个表、100万+ 数据到测试环境
传统方式：mysqldump 导出 → 传输文件 → 导入，耗时 2 小时
使用本工具：配置一次，执行命令，边喝咖啡边看进度，30 分钟完成
```

**场景 2：跨区域同步（带宽有限）**
```
问题：总部到分支网络带宽只有 2Mbps，大批量传输经常失败
传统方式：分表导出导入，手动操作繁琐，容易遗漏
使用本工具：配置 batch_size: 100，自动分批传输，稳定可靠
```

**场景 3：定期数据备份**
```
问题：每天需要将生产数据备份到备份服务器
传统方式：写 cron 脚本，维护复杂的 shell 命令
使用本工具：一行配置 + 一行 crontab，自动化完成
```

## 📦 快速开始

### 系统要求

- Go 1.21+ （仅编译时需要）
- MySQL 5.5+ / MariaDB 10.0+
- 网络连接到源数据库和目标数据库

### 安装方式

**方式一：从源码编译（推荐）**

```bash
# 1. 克隆项目
git clone https://github.com/yourusername/db-sync-tools.git
cd db-sync-tools

# 2. 编译（生成单个可执行文件）
go build -o db-sync cmd/db-sync/main.go

# 3. 完成！现在你有了一个 8MB 的可执行文件
```

**方式二：使用 Makefile**

```bash
make build          # 编译当前平台
make build-linux    # 编译 Linux 版本
make build-windows  # 编译 Windows 版本
```

### 三步开始使用

**第一步：创建配置文件**

```bash
# 复制示例配置
cp config.example.yaml config.yaml
```

**第二步：编辑配置（只需 3 个关键参数）**

```yaml
# 源数据库（从哪里复制数据）
source:
  url: "192.168.1.100:3306/production_db"  # 数据库地址
  username: "root"                          # 用户名
  password: "password123"                   # 密码

# 目标数据库（复制到哪里）
target:
  url: "192.168.1.200:3306/test_db"       # 数据库地址
  username: "root"                         # 用户名
  password: "password456"                  # 密码

# 同步设置（可选，使用默认值也很好）
sync:
  batch_size: 200              # 每批传输记录数
  truncate_before_sync: false  # 是否清空目标表
```

**第三步：运行同步**

```bash
# 执行同步
./db-sync -config config.yaml

# 看到实时进度
# [INFO] 开始同步表: users (总行数: 10000)
# [INFO] 表 users: 批次 25/50, 进度: 50.00% (5000/10000), 速度: 200 行/秒
# [INFO] 表 users 同步完成: 共同步 10000 行, 耗时 50s
```

就这么简单！☕️

## ⚙️ 配置说明

### 完整配置说明

```yaml
# 源数据库配置
source:
  url: "192.168.1.100:3306/source_db"  # 必填：host:port/database
  username: "root"                      # 必填：数据库用户名
  password: "password"                  # 必填：数据库密码

# 目标数据库配置
target:
  url: "192.168.1.200:3306/target_db"  # 必填：host:port/database
  username: "root"                      # 必填：数据库用户名
  password: "password"                  # 必填：数据库密码

# 同步配置（可选）
sync:
  batch_size: 200              # 每批传输记录数，默认 200
  truncate_before_sync: false  # 是否清空目标表，默认 false
  
  # 排除某些表（支持通配符）
  exclude_tables:
    - "tmp_*"      # 排除所有 tmp_ 开头的表
    - "log_*"      # 排除所有 log_ 开头的表
    - "cache_*"    # 排除所有 cache_ 开头的表
  
  # 只同步指定表（如果为空则同步所有）
  include_tables: []
  
  timeout: 3600    # 超时时间（秒），默认 3600
  verbose: true    # 详细日志，默认 true

# 日志配置（可选）
log:
  level: "info"       # 日志级别：debug, info, warn, error
  file: "sync.log"    # 日志文件路径
  console: true       # 是否输出到控制台
```

### 常用命令

```bash
# 基本使用
./db-sync                              # 使用默认配置 config.yaml
./db-sync -config my-config.yaml       # 使用指定配置文件
./db-sync -version                     # 查看版本信息

# 配合其他命令
./db-sync -config config.yaml &        # 后台运行
nohup ./db-sync -config config.yaml &  # 后台运行（断开SSH后继续）
```

## 📖 使用场景

### 场景 1：本地测试环境搭建

```yaml
source:
  url: "production.example.com:3306/app_db"
  username: "readonly_user"
  password: "safe_password"

target:
  url: "localhost:3306/dev_db"
  username: "root"
  password: "local_password"

sync:
  batch_size: 500
  truncate_before_sync: true    # 完全覆盖本地数据
  exclude_tables:
    - "logs_*"                  # 排除日志表
    - "sessions"                # 排除会话表
```

### 场景 2：跨区域低带宽同步

```yaml
source:
  url: "10.1.1.100:3306/main_db"
  username: "sync_user"
  password: "password"

target:
  url: "10.2.1.100:3306/backup_db"
  username: "sync_user"
  password: "password"

sync:
  batch_size: 50              # 小批量，适应低带宽
  timeout: 7200               # 延长超时时间
```

### 场景 3：只同步特定表

```yaml
source:
  url: "192.168.1.100:3306/app_db"
  username: "root"
  password: "password"

target:
  url: "192.168.1.200:3306/app_db"
  username: "root"
  password: "password"

sync:
  include_tables:             # 只同步这些表
    - "users"
    - "orders"
    - "products"
  truncate_before_sync: true
```

### 场景 4：定时自动同步

创建 crontab 任务（Linux）：

```bash
# 每天凌晨 2 点同步
0 2 * * * cd /opt/db-sync-tools && ./db-sync -config config.yaml >> sync-cron.log 2>&1
```

Windows 任务计划程序：

```batch
# 创建批处理文件 sync.bat
cd C:\db-sync-tools
db-sync.exe -config config.yaml
```

然后在任务计划程序中添加定时任务。

## 💡 配置参数详解

### batch_size（批量大小）

**作用**：控制每次传输的记录数

**建议值**：
- 🌐 低带宽（<1Mbps）：50-100
- 🌐 中等带宽（1-10Mbps）：200-500  ← 默认值
- 🌐 高带宽（>10Mbps）：1000-5000
- 💾 有大字段（BLOB/TEXT）：10-50

### truncate_before_sync（清空目标表）

**选项**：
- ✅ `true`：同步前清空目标表（**完全覆盖模式**）
  - 适合：测试环境搭建、数据备份、完全同步
  - 风险：目标表原有数据将被清空
  
- ❌ `false`：保留目标表数据（**追加模式**）
  - 适合：增量同步、数据合并
  - 风险：可能出现主键冲突

### exclude_tables / include_tables（表过滤）

**排除表**（黑名单）：
```yaml
exclude_tables:
  - "tmp_*"        # 通配符：所有 tmp_ 开头的表
  - "log_2024*"    # 通配符：所有 log_2024 开头的表
  - "test_table"   # 精确匹配：指定表名
```

**包含表**（白名单）：
```yaml
include_tables:
  - "users"        # 只同步这些表
  - "orders"
  - "products"
```

⚠️ **注意**：如果 `include_tables` 不为空，只会同步指定的表，`exclude_tables` 在此基础上生效。

## 项目结构

```
db-sync-tools/
├── cmd/
│   └── db-sync/
│       └── main.go              # 主程序入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理模块
│   ├── database/
│   │   └── database.go          # 数据库连接模块
│   ├── logger/
│   │   └── logger.go            # 日志模块
│   ├── metadata/
│   │   └── metadata.go          # 表元数据模块
│   └── sync/
│       └── sync.go              # 数据同步核心模块
├── config.example.yaml          # 配置文件示例
├── go.mod                       # Go 模块文件
├── go.sum                       # Go 依赖锁定文件
├── README.md                    # 项目说明文档
├── Makefile                     # 构建脚本
└── .gitignore                   # Git 忽略文件

```

## 模块说明

### 1. config 模块
- 负责配置文件的加载和验证
- 支持 YAML 格式配置
- 提供配置项验证功能

### 2. database 模块
- 封装 MySQL 数据库连接
- 提供表操作的基础方法
- 管理连接池

### 3. logger 模块
- 提供多级别日志记录（DEBUG, INFO, WARN, ERROR）
- 支持控制台和文件双输出
- 彩色日志输出

### 4. metadata 模块
- 表结构发现和元数据收集
- 表过滤（包含/排除）
- 主键和列信息获取

### 5. sync 模块
- 核心数据同步逻辑
- 批量读取和写入
- 进度统计和速度计算
- 错误处理和重试

## 工作原理

1. **连接数据库**：建立到源数据库和目标数据库的连接
2. **发现表**：获取源数据库中的所有表，根据配置过滤
3. **收集元数据**：获取每个表的结构、行数、主键等信息
4. **创建表结构**：在目标数据库中创建相同结构的表（如果不存在）
5. **批量同步**：
   - 使用 `LIMIT` 和 `OFFSET` 分批读取源数据
   - 使用批量 `INSERT` 写入目标数据库
   - 实时显示进度和速度
6. **统计报告**：同步完成后显示详细的统计信息

## 性能优化建议

1. **调整批量大小**：根据网络带宽和数据大小调整 `batch_size`
2. **网络优化**：确保源数据库和目标数据库之间的网络连接稳定
3. **索引处理**：考虑在同步前临时禁用目标表的索引，同步后重建
4. **并行同步**：对于多表同步，可以考虑修改代码支持并行处理
5. **增量同步**：使用时间戳或ID字段实现增量同步（需要修改代码）

## 注意事项

⚠️ **重要提示**：

1. 建议先在测试环境中验证同步功能
2. 同步大型数据库前请确保目标数据库有足够的存储空间
3. 如果启用 `truncate_before_sync`，目标表的数据将被清空
4. 对于生产环境，建议在业务低峰期执行同步
5. 同步过程中请勿中断，否则可能导致数据不一致
6. 建议定期备份数据库

## ❓ 常见问题

### Q1: 同步速度很慢怎么办？

**原因分析**：
- 网络带宽不足
- batch_size 设置太小
- 数据库服务器性能瓶颈

**解决方法**：
```yaml
sync:
  batch_size: 1000  # 适当增大批量（前提是网络和内存充足）
```

### Q2: 出现主键冲突错误

**错误信息**：`Error 1062: Duplicate entry '1' for key 'PRIMARY'`

**原因**：目标表已有数据，且未清空

**解决方法**：
```yaml
sync:
  truncate_before_sync: true  # 启用清空目标表
```

或手动清空：
```sql
TRUNCATE TABLE table_name;
```

### Q3: 表没有主键能同步吗？

**回答**：✅ 可以！

本工具支持无主键的表，会显示警告但继续同步：
```
[WARN] 获取表 my_table 的主键失败
[INFO] 开始同步表: my_table (总行数: 1000)
```

### Q4: 如何中断正在运行的同步？

**方法**：按 `Ctrl+C`

程序会优雅退出并显示已完成的统计信息：
```
[WARN] 收到信号: interrupt，正在优雅退出...
[INFO] 总表数: 50, 成功: 10, 失败: 0
```

### Q5: 支持哪些 MySQL 版本？

**支持版本**：
- ✅ MySQL 5.5, 5.6, 5.7, 8.0
- ✅ MariaDB 10.0+
- ✅ Percona Server

**自动适配**：程序会自动检测并适配不同版本。

### Q6: 内存占用过高怎么办？

**解决方法**：
```yaml
sync:
  batch_size: 50  # 减小批量大小
```

特别是表中有大字段（BLOB、TEXT）时，建议使用较小的批量。

### Q7: 连接数据库失败

**常见错误**：
```
连接数据库失败: Error 1045: Access denied
```

**检查清单**：
- ✅ 用户名和密码是否正确
- ✅ 用户是否有足够权限（源需要 SELECT，目标需要 INSERT、CREATE、DROP）
- ✅ 数据库服务是否运行
- ✅ 防火墙是否允许连接
- ✅ URL 格式是否正确（`host:port/database`）

**测试连接**：
```bash
mysql -h host -P port -u username -p database
```

### Q8: 如何只同步表结构不同步数据？

**当前版本**：不支持

**替代方案**：
```bash
# 使用 mysqldump 只导出结构
mysqldump -h host -u user -p --no-data database > schema.sql
```

## 🔧 故障排除速查表

| 问题 | 可能原因 | 解决方法 |
|------|----------|----------|
| 连接失败 | 网络/权限/配置 | 检查 URL、用户名、密码、防火墙 |
| 速度慢 | batch_size 太小 | 增大 batch_size 到 500-1000 |
| 主键冲突 | 目标表有数据 | 启用 truncate_before_sync |
| 内存过高 | batch_size 太大 | 减小 batch_size 到 50-100 |
| Ctrl+C 无效 | 正在执行批次 | 等待当前批次完成或减小 batch_size |
| 表结构不同 | 手动修改过表 | 确保源和目标表结构一致 |

## 日志

日志文件默认保存在 `sync.log`，包含：

- 连接信息
- 表同步进度
- 错误信息
- 统计数据

日志级别：

- `debug`：详细的调试信息
- `info`：一般信息（默认）
- `warn`：警告信息
- `error`：错误信息

## 开发

### 添加新功能

项目采用模块化设计，可以方便地扩展功能：

1. 在 `internal/` 目录下创建新模块
2. 在 `cmd/db-sync/main.go` 中集成新模块
3. 更新配置文件和文档

### 运行测试

```bash
go test ./...
```

### 代码格式化

```bash
go fmt ./...
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 作者

Your Name

## 🎯 核心特性一览

| 特性 | 说明 | 优势 |
|------|------|------|
| 🚀 批量传输 | 可配置批量大小 | 适应各种网络环境 |
| 📊 实时进度 | 显示进度、速度 | 随时掌握同步状态 |
| ⚙️ 简单配置 | 3 个参数即可 | 5 分钟快速上手 |
| 🛡️ 稳定可靠 | 支持所有 MySQL 版本 | 兼容性强 |
| 🎨 表过滤 | 包含/排除表 | 灵活控制同步范围 |
| 💾 无主键支持 | 自动适配 | 无主键表也能同步 |
| ⏸️ 优雅退出 | Ctrl+C 随时中断 | 安全可控 |
| 📝 详细日志 | 多级别日志 | 便于问题诊断 |

## 🏗️ 技术架构

### 模块化设计

```
db-sync-tools/
├── cmd/db-sync/          主程序入口
├── internal/
│   ├── config/           配置管理（3参数简化）
│   ├── database/         数据库连接（动态适配）
│   ├── logger/           日志系统（多级别）
│   ├── metadata/         元数据管理（表发现）
│   └── sync/             同步引擎（批量处理）
└── config.yaml           配置文件
```

### 工作流程

```
1. 读取配置 → 2. 连接数据库 → 3. 发现表 → 4. 收集元数据
                                                    ↓
8. 统计报告 ← 7. 实时进度 ← 6. 批量插入 ← 5. 批量读取
```

### 性能数据

| 环境 | 批量大小 | 速度（行/秒） |
|------|---------|--------------|
| 100M 带宽 | 200 | 500-1000 |
| 1G 带宽 | 1000 | 5000-10000 |
| 局域网 | 5000 | 10000-50000 |

*实际速度受多种因素影响：网络、数据大小、数据库性能等*

## 📚 更多文档

- 📖 [完整文档](docs/) - 详细的使用文档
- 🔍 [使用示例](docs/EXAMPLES.md) - 10+ 个实际场景
- 🐛 [问题反馈](https://github.com/yourusername/db-sync-tools/issues) - 提交 Bug 或建议

## 🤝 贡献指南

欢迎贡献代码、文档或提出建议！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 💬 联系方式

- 📧 Email: your.email@example.com
- 🐛 Issues: https://github.com/yourusername/db-sync-tools/issues
- 💡 Discussions: https://github.com/yourusername/db-sync-tools/discussions

## ⭐ 如果这个项目对你有帮助，请给个 Star！

---

**最后更新**: 2025-11-04  
**版本**: v1.0.2  
**状态**: ✅ 稳定版，生产可用

