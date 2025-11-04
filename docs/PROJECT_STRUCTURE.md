# 项目架构说明

本文档详细说明了 MySQL 数据库同步工具的架构设计和模块划分。

## 架构设计原则

1. **模块化**: 每个功能模块独立，职责明确
2. **可扩展**: 易于添加新功能和模块
3. **可维护**: 代码结构清晰，便于维护
4. **高性能**: 批量处理，优化网络传输
5. **可靠性**: 完善的错误处理和日志记录

## 项目目录结构

```
db-sync-tools/
│
├── cmd/                          # 命令行应用程序
│   └── db-sync/
│       └── main.go              # 主程序入口点
│
├── internal/                     # 内部包（不对外暴露）
│   ├── config/                  # 配置管理模块
│   │   └── config.go
│   │
│   ├── database/                # 数据库操作模块
│   │   └── database.go
│   │
│   ├── logger/                  # 日志记录模块
│   │   └── logger.go
│   │
│   ├── metadata/                # 元数据管理模块
│   │   └── metadata.go
│   │
│   └── sync/                    # 数据同步核心模块
│       └── sync.go
│
├── config.example.yaml          # 配置文件示例
├── go.mod                       # Go 模块定义
├── go.sum                       # 依赖版本锁定
├── Makefile                     # 构建脚本（Unix/Linux）
├── build.sh                     # 构建脚本（Shell）
├── build.bat                    # 构建脚本（Windows）
├── README.md                    # 项目说明文档
├── QUICKSTART.md                # 快速开始指南
├── EXAMPLES.md                  # 使用示例
├── PROJECT_STRUCTURE.md         # 项目架构说明（本文档）
└── .gitignore                   # Git 忽略文件
```

## 模块详细说明

### 1. cmd/db-sync（主程序模块）

**职责**: 
- 程序入口点
- 命令行参数解析
- 各模块的协调和调用
- 信号处理（优雅退出）

**关键功能**:
- `main()`: 程序入口
- `run()`: 主业务逻辑
- 信号处理: 支持 Ctrl+C 优雅退出

**依赖**:
- config（配置）
- database（数据库连接）
- logger（日志）
- metadata（元数据）
- sync（同步）

---

### 2. internal/config（配置模块）

**职责**:
- 读取和解析 YAML 配置文件
- 配置参数验证
- 提供配置数据结构

**数据结构**:

```go
type Config struct {
    Source DatabaseConfig  // 源数据库配置
    Target DatabaseConfig  // 目标数据库配置
    Sync   SyncConfig     // 同步配置
    Log    LogConfig      // 日志配置
}

type DatabaseConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    Database string
    Charset  string
}

type SyncConfig struct {
    BatchSize          int
    TruncateBeforeSync bool
    ExcludeTables      []string
    IncludeTables      []string
    Timeout            int
    Verbose            bool
}
```

**关键方法**:
- `Load(path string)`: 加载配置文件
- `Validate()`: 验证配置有效性
- `GetDSN()`: 生成数据库连接字符串

---

### 3. internal/database（数据库模块）

**职责**:
- 封装 MySQL 数据库连接
- 提供表操作的基础方法
- 管理连接池

**数据结构**:

```go
type DB struct {
    *sql.DB
    Config *config.DatabaseConfig
}

type ColumnInfo struct {
    Field   string
    Type    string
    Null    string
    Key     string
    Default string
    Extra   string
}
```

**关键方法**:
- `Connect(cfg)`: 建立数据库连接
- `GetTables()`: 获取所有表名
- `GetTableColumns(table)`: 获取表结构
- `GetRowCount(table)`: 获取表行数
- `GetPrimaryKey(table)`: 获取主键
- `TruncateTable(table)`: 清空表
- `TableExists(table)`: 检查表是否存在
- `CreateTableLike(table, ddl)`: 创建表结构
- `GetCreateTableSQL(table)`: 获取建表 SQL

---

### 4. internal/logger（日志模块）

**职责**:
- 多级别日志记录
- 控制台和文件双输出
- 彩色日志输出

**日志级别**:
- DEBUG: 详细调试信息
- INFO: 一般信息
- WARN: 警告信息
- ERROR: 错误信息

**关键方法**:
- `Init(level, file, console)`: 初始化日志系统
- `Debug(format, args...)`: 记录调试日志
- `Info(format, args...)`: 记录信息日志
- `Warn(format, args...)`: 记录警告日志
- `Error(format, args...)`: 记录错误日志
- `Fatal(format, args...)`: 记录致命错误并退出
- `Close()`: 关闭日志文件

**日志格式**:
```
2025-11-04 10:30:00 [INFO] 消息内容
```

---

### 5. internal/metadata（元数据模块）

**职责**:
- 发现和收集表元数据
- 表过滤（包含/排除）
- 提供表结构信息

**数据结构**:

```go
type TableMetadata struct {
    Name       string
    RowCount   int64
    Columns    []database.ColumnInfo
    PrimaryKey []string
    CreateSQL  string
}

type MetadataManager struct {
    sourceDB      *database.DB
    excludeTables []string
    includeTables []string
}
```

**关键方法**:
- `NewMetadataManager()`: 创建元数据管理器
- `DiscoverTables()`: 发现需要同步的表
- `GetTableMetadata(table)`: 获取表的元数据
- `shouldSyncTable(table)`: 判断表是否应该同步
- `matchPattern(name, pattern)`: 模式匹配（支持通配符）

**表过滤逻辑**:
1. 如果 `include_tables` 不为空，只同步包含的表
2. 否则同步所有表，但排除 `exclude_tables` 中的表
3. 支持通配符匹配（*, ?）

---

### 6. internal/sync（同步核心模块）

**职责**:
- 核心数据同步逻辑
- 批量读取和写入
- 进度统计和速度计算
- 错误处理

**数据结构**:

```go
type Syncer struct {
    sourceDB           *database.DB
    targetDB           *database.DB
    batchSize          int
    truncateBeforeSync bool
    timeout            time.Duration
}

type SyncStats struct {
    TotalTables   int
    SyncedTables  int
    FailedTables  int
    TotalRows     int64
    SyncedRows    int64
    StartTime     time.Time
    EndTime       time.Time
    Duration      time.Duration
    FailedDetails []FailedTableInfo
}
```

**关键方法**:
- `NewSyncer()`: 创建同步器
- `SyncTable(ctx, meta)`: 同步单个表
- `SyncAll(ctx, tables, manager)`: 同步所有表
- `readBatch(table, columns, offset, limit)`: 读取一批数据
- `writeBatch(meta, rows)`: 写入一批数据

**同步流程**:
1. 检查目标表是否存在，不存在则创建
2. 如果需要，清空目标表
3. 分批读取源表数据（使用 LIMIT/OFFSET）
4. 批量插入目标表
5. 实时显示进度和速度
6. 记录统计信息

---

## 数据流图

```
┌─────────────┐
│   main.go   │  启动程序
└──────┬──────┘
       │
       ├──> config.Load()          读取配置
       │
       ├──> logger.Init()          初始化日志
       │
       ├──> database.Connect()     连接源数据库
       │    database.Connect()     连接目标数据库
       │
       ├──> metadata.DiscoverTables()  发现表
       │    metadata.GetTableMetadata()  获取元数据
       │
       └──> sync.SyncAll()         执行同步
            │
            └─> For each table:
                ├─> sync.SyncTable()
                │   ├─> readBatch()     读取数据
                │   └─> writeBatch()    写入数据
                │
                └─> 统计和报告
```

## 技术栈

- **语言**: Go 1.21+
- **数据库驱动**: github.com/go-sql-driver/mysql
- **配置解析**: gopkg.in/yaml.v3
- **标准库**: 
  - `database/sql`: 数据库操作
  - `context`: 上下文控制
  - `flag`: 命令行参数
  - `os/signal`: 信号处理

## 性能优化

### 1. 批量处理
- 使用批量 INSERT 而非逐行插入
- 可配置的批量大小（默认 200）
- 减少网络往返次数

### 2. 连接池管理
```go
db.SetMaxOpenConns(10)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(time.Hour)
```

### 3. 内存优化
- 流式读取，不一次性加载所有数据
- 及时释放资源
- 使用 LIMIT/OFFSET 分页

### 4. 并发控制
- 当前版本：顺序同步表
- 可扩展：支持并行同步多个表

## 错误处理策略

### 1. 连接错误
- 详细的错误信息
- 自动重试机制（待实现）
- 连接健康检查

### 2. 数据错误
- 记录失败的表
- 继续处理其他表
- 最终报告所有失败

### 3. 网络错误
- 超时控制
- 优雅退出
- 断点续传（待实现）

## 安全考虑

### 1. 配置文件安全
- 配置文件不纳入版本控制
- 建议使用环境变量存储密码
- 限制配置文件权限

### 2. 数据库权限
- 源数据库: 只需 SELECT 权限
- 目标数据库: 需要 SELECT, INSERT, CREATE, DROP 权限

### 3. 网络安全
- 支持 SSL 连接（配置 DSN）
- VPN 隧道（外部实现）

## 扩展性设计

### 可扩展点

1. **新的数据库支持**
   - 实现 `database.DB` 接口
   - PostgreSQL, SQL Server 等

2. **过滤器**
   - 自定义表过滤逻辑
   - 行级过滤（WHERE 条件）

3. **转换器**
   - 数据类型转换
   - 数据清洗和转换

4. **通知机制**
   - 邮件通知
   - Webhook 通知
   - 消息队列

5. **监控指标**
   - Prometheus metrics
   - 性能监控
   - 告警系统

### 未来功能规划

- [ ] 增量同步（基于时间戳）
- [ ] 断点续传
- [ ] 并行表同步
- [ ] 数据校验（checksum）
- [ ] 双向同步
- [ ] Web UI 控制台
- [ ] RESTful API
- [ ] 分布式部署

## 测试策略

### 单元测试
```bash
go test ./internal/config
go test ./internal/database
go test ./internal/metadata
go test ./internal/sync
```

### 集成测试
- 测试数据库环境准备
- 完整同步流程测试
- 异常场景测试

### 性能测试
- 大数据量测试
- 并发测试
- 压力测试

## 部署建议

### 1. 单机部署
```bash
# 编译
go build -o db-sync cmd/db-sync/main.go

# 配置
cp config.example.yaml config.yaml
vim config.yaml

# 运行
./db-sync -config config.yaml
```

### 2. Docker 部署
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o db-sync cmd/db-sync/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/db-sync .
COPY config.yaml .
CMD ["./db-sync", "-config", "config.yaml"]
```

### 3. Kubernetes 部署
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: db-sync
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: db-sync
            image: db-sync:latest
            volumeMounts:
            - name: config
              mountPath: /app/config.yaml
          volumes:
          - name: config
            configMap:
              name: db-sync-config
```

## 贡献指南

### 代码规范
- 遵循 Go 官方代码规范
- 使用 `go fmt` 格式化代码
- 添加必要的注释
- 编写单元测试

### 提交流程
1. Fork 项目
2. 创建功能分支
3. 提交代码
4. 编写测试
5. 发起 Pull Request

## 许可证

MIT License

---

如有问题，欢迎提交 Issue 或 Pull Request！

