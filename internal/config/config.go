package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	// 多数据源集合，key 为数据源别名
	DataSources map[string]DatabaseConfig `yaml:"datasources"`
	// 同步与日志配置
	Sync SyncConfig `yaml:"sync"`
	Log  LogConfig  `yaml:"log"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	URL      string `yaml:"url"`      // 数据库连接地址，格式: host:port/database
	Username string `yaml:"username"` // 数据库用户名
	Password string `yaml:"password"` // 数据库密码
}

// SyncConfig 同步配置
type SyncConfig struct {
	// 使用别名指定源与目标数据源
	SourceAlias        string   `yaml:"source"`
	TargetAlias        string   `yaml:"target"`
	BatchSize          int      `yaml:"batch_size"`
	Concurrency        int      `yaml:"concurrency"`
	TruncateBeforeSync bool     `yaml:"truncate_before_sync"`
	ExcludeTables      []string `yaml:"exclude_tables"`
	IncludeTables      []string `yaml:"include_tables"`
	Timeout            int      `yaml:"timeout"`
	Verbose            bool     `yaml:"verbose"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level   string `yaml:"level"`
	File    string `yaml:"file"`
	Console bool   `yaml:"console"`
}

// Load 从文件加载配置
func Load(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 校验数据源集合
	if len(c.DataSources) == 0 {
		return fmt.Errorf("必须配置至少一个数据源(datasources)")
	}
	for alias, ds := range c.DataSources {
		if err := ds.Validate(fmt.Sprintf("数据源[%s]", alias)); err != nil {
			return err
		}
	}

	// 验证同步配置
	if c.Sync.BatchSize <= 0 {
		return fmt.Errorf("批量大小必须大于0")
	}
	if c.Sync.Concurrency <= 0 {
		c.Sync.Concurrency = 1
	}
	if c.Sync.Timeout <= 0 {
		c.Sync.Timeout = 3600 // 默认1小时
	}
	if c.Sync.SourceAlias == "" {
		return fmt.Errorf("sync.source 不能为空（源数据源别名）")
	}
	if c.Sync.TargetAlias == "" {
		return fmt.Errorf("sync.target 不能为空（目标数据源别名）")
	}
	if _, ok := c.DataSources[c.Sync.SourceAlias]; !ok {
		return fmt.Errorf("未找到源数据源别名: %s", c.Sync.SourceAlias)
	}
	if _, ok := c.DataSources[c.Sync.TargetAlias]; !ok {
		return fmt.Errorf("未找到目标数据源别名: %s", c.Sync.TargetAlias)
	}

	return nil
}

// Validate 验证数据库配置
func (d *DatabaseConfig) Validate(name string) error {
	if d.URL == "" {
		return fmt.Errorf("%s: URL 不能为空", name)
	}
	if d.Username == "" {
		return fmt.Errorf("%s: 用户名不能为空", name)
	}
	// 密码可以为空（某些情况下）
	return nil
}

// GetDSN 获取数据库连接字符串
func (d *DatabaseConfig) GetDSN() string {
	// URL 格式: host:port/database
	// DSN 格式: username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local

	// 查找最后一个 / 的位置，分离 host:port 和 database
	lastSlash := strings.LastIndex(d.URL, "/")
	if lastSlash == -1 {
		// 如果没有 /，则整个 URL 是 host:port，没有指定数据库
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

// LoadOrDefault 加载配置文件，如果不存在则使用默认配置
func LoadOrDefault(configPath string) (*Config, error) {
	// 如果文件不存在，创建示例配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 检查是否存在示例配置
		examplePath := "config.example.yaml"
		if _, err := os.Stat(examplePath); err == nil {
			return nil, fmt.Errorf("配置文件 %s 不存在，请参考 %s 创建配置文件", configPath, examplePath)
		}
		return nil, fmt.Errorf("配置文件 %s 不存在", configPath)
	}

	return Load(configPath)
}

// GetSourceConfig 返回源数据源配置
func (c *Config) GetSourceConfig() *DatabaseConfig {
	if ds, ok := c.DataSources[c.Sync.SourceAlias]; ok {
		return &ds
	}
	return nil
}

// GetTargetConfig 返回目标数据源配置
func (c *Config) GetTargetConfig() *DatabaseConfig {
	if ds, ok := c.DataSources[c.Sync.TargetAlias]; ok {
		return &ds
	}
	return nil
}

// GetAbsPath 获取相对于配置文件的绝对路径
func GetAbsPath(basePath, relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	baseDir := filepath.Dir(basePath)
	return filepath.Join(baseDir, relativePath)
}
