package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config Open API 配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port            string `mapstructure:"port"`
	Mode            string `mapstructure:"mode"`
	ReadTimeout     int    `mapstructure:"read_timeout"`
	WriteTimeout    int    `mapstructure:"write_timeout"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// StorageConfig 文件存储配置
type StorageConfig struct {
	UploadDir         string   `mapstructure:"upload_dir"`
	MaxUploadSizeMB   int64    `mapstructure:"max_upload_size_mb"`
	AllowedExtensions []string `mapstructure:"allowed_extensions"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	RequestsPerMin int  `mapstructure:"requests_per_minute"`
}

// 全局配置实例
var globalConfig *Config

// Load 加载配置
func Load(configPaths ...string) (*Config, error) {
	v := viper.New()

	// 设置配置名和类型
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 添加配置搜索路径（优先级从低到高）
	searchPaths := []string{
		".",
		"./config",
		"/etc/contful/",
		"$HOME/.contful/",
	}
	// 添加自定义路径
	if len(configPaths) > 0 {
		searchPaths = append(configPaths, searchPaths...)
	}
	for _, path := range searchPaths {
		v.AddConfigPath(path)
	}

	// 环境变量支持
	v.SetEnvPrefix("CONTFUL_OPEN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取环境变量覆盖
	readEnvOverrides(v)

	// 设置默认值
	setDefaults(v)

	// 尝试读取配置文件（忽略不存在的错误）
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 解析配置到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 处理字符串切片默认值
	if len(cfg.Storage.AllowedExtensions) == 0 {
		cfg.Storage.AllowedExtensions = []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp",
			".pdf", ".doc", ".docx", ".xls", ".xlsx",
		}
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 后处理
	cfg.PostLoad()

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if c.Storage.UploadDir == "" {
		c.Storage.UploadDir = "./uploads"
	}
	return nil
}

// PostLoad 后处理
func (c *Config) PostLoad() {
	// 服务默认值
	if c.Server.Port == "" {
		c.Server.Port = "8081"
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "release"
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 30
	}

	// 数据库默认值
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 100
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 10
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 3600
	}

	// Redis 默认值
	if c.Redis.Port == 0 {
		c.Redis.Port = 6379
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 100
	}

	// 存储默认值
	if c.Storage.MaxUploadSizeMB == 0 {
		c.Storage.MaxUploadSizeMB = 10
	}

	// 日志默认值
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}

	// 限流默认值
	if c.RateLimit.RequestsPerMin == 0 {
		c.RateLimit.RequestsPerMin = 100
	}
}

// GetDSN 获取 PostgreSQL DSN
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetAddr 获取 Redis 地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetMaxUploadSize 获取最大上传大小(字节)
func (c *StorageConfig) GetMaxUploadSize() int64 {
	return c.MaxUploadSizeMB * 1024 * 1024
}

// GetConnMaxLifetime 获取连接最大存活时间
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// IsExtensionAllowed 检查扩展名是否允许
func (c *StorageConfig) IsExtensionAllowed(filename string) bool {
	ext := getExt(filename)
	for _, allowed := range c.AllowedExtensions {
		if strings.EqualFold(ext, allowed) {
			return true
		}
	}
	return false
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// 服务
	v.SetDefault("server.port", "8081")
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.read_timeout", 60)
	v.SetDefault("server.write_timeout", 60)
	v.SetDefault("server.shutdown_timeout", 30)

	// 数据库
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 3600)

	// Redis
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)

	// 存储
	v.SetDefault("storage.upload_dir", "./uploads")
	v.SetDefault("storage.max_upload_size_mb", 10)

	// 日志
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// 限流
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.requests_per_minute", 100)
}

// readEnvOverrides 读取环境变量覆盖
func readEnvOverrides(v *viper.Viper) {
	envMappings := map[string]string{
		"DB_HOST":         "database.host",
		"DB_PORT":         "database.port",
		"DB_USER":         "database.user",
		"DB_PASSWORD":     "database.password",
		"DB_NAME":         "database.name",
		"DB_SSL_MODE":     "database.ssl_mode",
		"REDIS_HOST":      "redis.host",
		"REDIS_PORT":      "redis.port",
		"REDIS_PASSWORD":  "redis.password",
		"REDIS_DB":        "redis.db",
		"SERVER_PORT":     "server.port",
	}

	for envKey, configKey := range envMappings {
		if val := os.Getenv(envKey); val != "" {
			v.Set(configKey, val)
		}
	}
}

// 辅助函数
func getExt(filename string) string {
	i := strings.LastIndex(filename, ".")
	if i == -1 {
		return ""
	}
	return filename[i:]
}
