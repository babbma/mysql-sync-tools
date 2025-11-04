# DSN 错误故障排除指南

## 问题描述

**错误信息**：
```
invalid DSN: network address not terminated (missing closing brace)
```

**完整错误示例**：
```
2025-11-04 09:23:43 [ERROR] 同步失败: 连接源数据库失败: 打开数据库连接失败: 
invalid DSN: network address not terminated (missing closing brace)
```

## 问题原因

这是一个 **DSN（数据源名称）格式错误**。MySQL 的 DSN 格式有严格的要求。

### DSN 格式说明

**正确的 DSN 格式**：
```
username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
```

**关键点**：
- `@tcp(host:port)` - TCP 连接地址在括号内
- `)/database?` - 数据库名在括号外，用 `/` 分隔
- 括号必须正确配对

### 错误示例

❌ **错误的格式**（旧版本代码）：
```
root:password@tcp(2.tcp.vip.cpolar.cn:10972/INSITE_changcheng)?charset=utf8mb4
                                              ↑
                                        数据库名不应该在括号内！
```

✅ **正确的格式**（修复后）：
```
root:password@tcp(2.tcp.vip.cpolar.cn:10972)/INSITE_changcheng?charset=utf8mb4
                                           ↑↑
                                    括号在这里闭合，数据库名在括号外
```

## 解决方案

### 方案 1: 更新到最新版本（推荐）

已在代码中修复此问题，重新编译即可：

```bash
# 重新编译
go build -o db-sync.exe cmd/db-sync/main.go

# 或使用构建脚本
build.bat  # Windows
./build.sh # Linux/Mac
```

### 方案 2: 验证配置文件

确保你的配置文件格式正确：

**正确的配置**：
```yaml
source:
  url: "2.tcp.vip.cpolar.cn:10972/INSITE_changcheng"
  username: "root"
  password: "your_password"
```

**URL 格式要求**：
- 格式：`host:port/database`
- 例子：`localhost:3306/mydb`
- 例子：`192.168.1.100:3306/testdb`
- 例子：`db.example.com:3306/production`

## 技术细节

### 代码修复说明

**修复前的代码**：
```go
func (d *DatabaseConfig) GetDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s)?charset=utf8mb4&parseTime=True&loc=Local",
        d.Username,
        d.Password,
        d.URL,  // 直接使用整个 URL（包含 /database）
    )
}
```

**修复后的代码**：
```go
func (d *DatabaseConfig) GetDSN() string {
    // 查找最后一个 / 的位置，分离 host:port 和 database
    lastSlash := strings.LastIndex(d.URL, "/")
    if lastSlash == -1 {
        // 如果没有 /，则整个 URL 是 host:port
        return fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8mb4&parseTime=True&loc=Local",
            d.Username,
            d.Password,
            d.URL,
        )
    }
    
    // 分离 host:port 和 database
    hostPort := d.URL[:lastSlash]
    database := d.URL[lastSlash+1:]
    
    return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        d.Username,
        d.Password,
        hostPort,
        database,
    )
}
```

### DSN 生成示例

**输入配置**：
```yaml
url: "2.tcp.vip.cpolar.cn:10972/INSITE_changcheng"
username: "root"
password: "mypassword"
```

**生成的 DSN**：
```
root:mypassword@tcp(2.tcp.vip.cpolar.cn:10972)/INSITE_changcheng?charset=utf8mb4&parseTime=True&loc=Local
```

**分解说明**：
1. `root:mypassword` - 用户名和密码
2. `@tcp(2.tcp.vip.cpolar.cn:10972)` - TCP 连接地址
3. `/INSITE_changcheng` - 数据库名
4. `?charset=utf8mb4&parseTime=True&loc=Local` - 连接参数

## 验证方法

### 方法 1: 使用 MySQL 命令行测试

```bash
mysql -h 2.tcp.vip.cpolar.cn -P 10972 -u root -p INSITE_changcheng
```

如果能连接成功，说明配置正确。

### 方法 2: 使用工具测试

重新编译后运行：
```bash
./db-sync.exe -config config.yaml
```

应该能看到正常的连接信息：
```
2025-11-04 09:30:00 [INFO] 连接源数据库: root@2.tcp.vip.cpolar.cn:10972/INSITE_changcheng
2025-11-04 09:30:01 [INFO] 源数据库连接成功
```

## 常见 URL 格式示例

### 标准配置
```yaml
# 本地数据库
url: "localhost:3306/mydb"

# 远程数据库
url: "192.168.1.100:3306/production_db"

# 域名
url: "db.example.com:3306/app_db"

# 非标准端口
url: "192.168.1.100:3307/testdb"

# Cpolar 内网穿透
url: "2.tcp.vip.cpolar.cn:10972/mydb"
```

### 特殊情况

**数据库名包含特殊字符**：
```yaml
# 下划线（推荐）
url: "localhost:3306/my_database"

# 中划线
url: "localhost:3306/my-database"

# 中文数据库名（不推荐，但支持）
url: "localhost:3306/我的数据库"
```

## 其他可能的 DSN 错误

### 1. 端口号错误
```
❌ url: "localhost:mysql/mydb"      # 端口号必须是数字
✅ url: "localhost:3306/mydb"
```

### 2. 缺少数据库名
```
❌ url: "localhost:3306"            # 缺少数据库名
✅ url: "localhost:3306/mydb"
```

### 3. 多余的斜杠
```
❌ url: "localhost:3306//mydb"      # 多余的斜杠
✅ url: "localhost:3306/mydb"
```

### 4. 协议前缀
```
❌ url: "tcp://localhost:3306/mydb" # 不需要协议前缀
✅ url: "localhost:3306/mydb"
```

## 快速检查清单

在运行同步前，请检查：

- [ ] 重新编译了最新版本的程序
- [ ] URL 格式正确：`host:port/database`
- [ ] 端口号是数字
- [ ] 包含了数据库名
- [ ] 用户名和密码正确
- [ ] 数据库服务正常运行
- [ ] 网络连接正常
- [ ] 防火墙允许连接

## 测试配置

使用以下最小配置测试：

```yaml
# test-config.yaml
source:
  url: "localhost:3306/test_db"
  username: "root"
  password: "root"

target:
  url: "localhost:3306/test_db_backup"
  username: "root"
  password: "root"

sync:
  batch_size: 10
  truncate_before_sync: true
  
log:
  level: "debug"
  console: true
```

运行测试：
```bash
./db-sync.exe -config test-config.yaml
```

## 获取帮助

如果问题仍然存在：

1. **检查日志文件**：`sync.log`
2. **提交 Issue**：https://github.com/yourusername/db-sync-tools/issues
3. **邮件支持**：your.email@example.com

提交问题时请包含：
- 完整的错误信息
- 配置文件（隐藏密码）
- 数据库版本
- 操作系统版本

## 总结

这个错误已经在最新版本中修复。只需：

1. ✅ 重新编译程序
2. ✅ 确保配置格式正确
3. ✅ 运行测试

就可以正常使用了！

---

**更新日期**: 2025-11-04  
**修复版本**: v1.0.1  
**状态**: ✅ 已修复

如有其他问题，欢迎反馈！

