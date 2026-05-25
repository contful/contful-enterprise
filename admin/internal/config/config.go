// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"

	"github.com/contful/contful/admin/internal/crypto"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Storage   StorageConfig   `mapstructure:"storage"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Audit     AuditConfig     `mapstructure:"audit"`
	MultiSite MultiSiteConfig `mapstructure:"multi_site"`
	Features  FeaturesConfig  `mapstructure:"features"`
	Security  SecurityConfig  `mapstructure:"security"`
	Schedule  ScheduleConfig  `mapstructure:"schedule"`
}

// SupportedAlgorithms 支持的加密算法
var SupportedAlgorithms = []string{crypto.AlgorithmAES, crypto.AlgorithmSM4}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// Secret 主密钥（统一配置）
	Secret string `mapstructure:"secret"`

	// Algorithm 加密算法（由 crypto_mode 自动选择，也可手动覆盖）
	Algorithm string `mapstructure:"algorithm"`

	// CryptoMode 加密模式：rsa（国际算法）或 sm（国密全套）
	// 默认 rsa：RSA-2048 + SHA-256 + AES-256-GCM
	// sm：SM2 + SM3 + SM4-GCM 全链路国密
	CryptoMode string `mapstructure:"crypto_mode"`

	// RSA 密钥对文件路径（用于前端登录密码加密传输）
	// 相对路径相对于配置文件目录（conf/）
	// crypto_mode=sm 时会自动生成 SM2 密钥对，RSA 路径被忽略
	RSAPublicKeyPath  string `mapstructure:"rsa_pubkey_path"`
	RSAPrivateKeyPath string `mapstructure:"rsa_privkey_path"`
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
	// Type 数据库类型（postgres）
	Type            string `mapstructure:"type"`
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

// JWTConfig JWT 认证配置
type JWTConfig struct {
	Secret                   string `mapstructure:"secret"`
	AccessTokenExpireMinutes int    `mapstructure:"access_token_expire_minutes"`
	RefreshTokenExpireDays   int    `mapstructure:"refresh_token_expire_days"`
}

// StorageConfig 文件存储配置
type StorageConfig struct {
	Driver            string             `mapstructure:"driver"`
	UploadDir         string             `mapstructure:"upload_dir"`
	MaxUploadSizeMB   int64              `mapstructure:"max_upload_size_mb"`
	BaseURL           string             `mapstructure:"base_url"`
	AllowedExtensions []string           `mapstructure:"allowed_extensions"`
	Oss               CloudStorageConfig `mapstructure:"oss"`
}

// CloudStorageConfig 云存储配置（阿里云 OSS / 腾讯云 COS / 华为云 OBS / AWS S3 / MinIO）
type CloudStorageConfig struct {
	Bucket         string `mapstructure:"bucket"`
	Endpoint       string `mapstructure:"endpoint"`
	Region         string `mapstructure:"region"`
	BaseURL        string `mapstructure:"base_url"`
	PathPrefix     string `mapstructure:"path_prefix"`
	ForcePathStyle bool   `mapstructure:"force_path_style"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// AuditConfig 审计日志配置
type AuditConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	LogAllRequests bool   `mapstructure:"log_all_requests"`
	SigningKey     string `mapstructure:"signing_key"` // 审计日志 HMAC-SHA256 签名密钥（自动派生）
}

// MultiSiteConfig 多站点配置
type MultiSiteConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	DefaultSiteID string `mapstructure:"default_site_id"`
}

// FeaturesConfig 特性开关
type FeaturesConfig struct {
	VersionHistory bool `mapstructure:"version_history"`
	APITokens      bool `mapstructure:"api_tokens"`
	MediaLibrary   bool `mapstructure:"media_library"`
}

// ScheduleConfig 定时排期配置
type ScheduleConfig struct {
	Enabled  bool `mapstructure:"enabled"`  // 是否启用定时发布/下架调度器
	Interval int  `mapstructure:"interval"` // 扫描间隔（秒），默认 30
}

// 全局配置实例
var globalConfig *Config

// 全局加密 Provider（由 PostLoad 初始化）
var globalCryptoProvider crypto.CryptoProvider

// Load 加载配置
func Load(configPaths ...string) (*Config, error) {
	v := viper.New()

	// 读取 .env 文件（如果存在）
	// 搜索路径：当前目录 → config/ 目录
	for _, p := range []string{".", "./config"} {
		envPath := filepath.Join(p, ".env")
		if _, err := os.Stat(envPath); err == nil {
			gotenv.OverLoad(envPath)
			break
		}
	}

	// 设置配置名和类型
	v.SetConfigName("console")
	v.SetConfigType("yaml")

	// 添加配置搜索路径（优先级从低到高）
	searchPaths := []string{
		".",
		"./conf",
		"../conf",
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

	// 读取环境变量覆盖（通过 readEnvOverrides 统一处理）
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

	// 处理字符串切片默认值（YAML 中可能为空）
	if len(cfg.Storage.AllowedExtensions) == 0 {
		cfg.Storage.AllowedExtensions = []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp",
			".pdf", ".doc", ".docx", ".xls", ".xlsx",
		}
	}
	if len(cfg.CORS.AllowedOrigins) == 0 {
		cfg.CORS.AllowedOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	}
	if len(cfg.CORS.AllowedMethods) == 0 {
		cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	// 验证必填配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 处理动态值
	cfg.PostLoad()

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// GetCryptoProvider 获取全局加密 Provider（对称 + 非对称 + 哈希一站式）
// 在 Load() 完成后可用，在此之前返回 nil
func GetCryptoProvider() crypto.CryptoProvider {
	return globalCryptoProvider
}

// Validate 验证配置
func (c *Config) Validate() error {
	// Secret 或 JWT Secret 至少配置一个
	if c.Security.Secret == "" && c.JWT.Secret == "" {
		return fmt.Errorf("security.secret or jwt.secret is required")
	}

	// JWT Secret 最小长度
	if len(c.JWT.Secret) > 0 && len(c.JWT.Secret) < 16 {
		return fmt.Errorf("jwt.secret must be at least 16 characters")
	}

	// 验证加密算法
	if c.Security.Algorithm != "" {
		valid := false
		for _, algo := range SupportedAlgorithms {
			if c.Security.Algorithm == algo {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported algorithm: %s, supported: %v", c.Security.Algorithm, SupportedAlgorithms)
		}
	}

	// 验证 crypto_mode（仅支持 rsa/sm 或空）
	if c.Security.CryptoMode != "" && c.Security.CryptoMode != crypto.ModeRSA && c.Security.CryptoMode != crypto.ModeSM {
		return fmt.Errorf("unsupported crypto_mode: %s, supported: %s / %s", c.Security.CryptoMode, crypto.ModeRSA, crypto.ModeSM)
	}

	// 数据库配置必填
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}

	// 上传目录必填
	if c.Storage.UploadDir == "" {
		c.Storage.UploadDir = "./uploads"
	}

	return nil
}

// PostLoad 后处理
func (c *Config) PostLoad() {
	// 确保默认端口
	if c.Server.Port == "" {
		c.Server.Port = "9080"
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "release"
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 30
	}

	// 数据库默认值
	if c.Database.Type == "" {
		c.Database.Type = "postgres" // 默认 PostgreSQL
	}
	if c.Database.Port == 0 {
		c.Database.Port = 5432 // PostgreSQL 默认端口
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

	// JWT 默认值
	if c.JWT.AccessTokenExpireMinutes == 0 {
		c.JWT.AccessTokenExpireMinutes = 15
	}
	if c.JWT.RefreshTokenExpireDays == 0 {
		c.JWT.RefreshTokenExpireDays = 7
	}

	// 文件存储默认值
	if c.Storage.MaxUploadSizeMB == 0 {
		c.Storage.MaxUploadSizeMB = 10
	}

	// CORS 默认值
	if c.CORS.MaxAge == 0 {
		c.CORS.MaxAge = 86400
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

	// Security 默认值
	if c.Security.Algorithm == "" {
		c.Security.Algorithm = "aes-256-gcm"
	}

	// 统一密钥派生：JWT/Token 加密使用 HKDF 从主密钥派生
	if c.Security.Secret != "" {
		// JWT 签名密钥：派生 "jwt" info
		if c.JWT.Secret == "" {
			c.JWT.Secret = deriveKey(c.Security.Secret, "jwt-signing", 32)
		}
	}

	// 审计签名密钥：派生 "audit-signing" info
	c.Audit.SigningKey = deriveKey(c.Security.Secret, "audit-signing", 32)

	// RSA 密钥对：从文件读取，未配置时自动生成
	// 公钥用于 PublicKey 端点，私钥用于密码解密
	if c.Security.RSAPublicKeyPath == "" {
		c.Security.RSAPublicKeyPath = "rsa_public.pem"
	}
	if c.Security.RSAPrivateKeyPath == "" {
		c.Security.RSAPrivateKeyPath = "rsa_private.pem"
	}

	// CryptoMode 默认值
	if c.Security.CryptoMode == "" {
		c.Security.CryptoMode = crypto.ModeRSA
	}

	// 定时排期默认值
	if c.Schedule.Interval == 0 {
		c.Schedule.Interval = 30
	}

	// 国密模式下自动切换对称加密算法
	if c.Security.CryptoMode == crypto.ModeSM {
		c.Security.Algorithm = crypto.AlgorithmSM4
	}

	// 初始化全局加密 Provider
	if c.Security.Secret != "" {
		provider, err := crypto.NewProvider(c.Security.CryptoMode, c.Security.Secret)
		if err != nil {
			// 日志警告但继续启动（Provider 为 nil 时调用方自行处理）
			globalCryptoProvider = nil
		} else {
			globalCryptoProvider = provider
		}
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

// GetConnMaxLifetime 获取连接最大存活时间
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// 服务
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.read_timeout", 60)
	v.SetDefault("server.write_timeout", 60)
	v.SetDefault("server.shutdown_timeout", 30)

	// 数据库
	v.SetDefault("database.type", "postgres")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 3600)

	// Redis
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)

	// JWT
	v.SetDefault("jwt.access_token_expire_minutes", 15)
	v.SetDefault("jwt.refresh_token_expire_days", 7)

	// 存储
	v.SetDefault("storage.driver", "local")
	v.SetDefault("storage.upload_dir", "./uploads")
	v.SetDefault("storage.max_upload_size_mb", 10)
	v.SetDefault("storage.base_url", "/assets")

	// CORS
	v.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization", "X-Requested-With"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", 86400)

	// 日志
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// 审计
	v.SetDefault("audit.enabled", true)
	v.SetDefault("audit.log_all_requests", false)

	// 特性开关
	v.SetDefault("features.version_history", true)
	v.SetDefault("features.api_tokens", true)
	v.SetDefault("features.media_library", true)

	// 定时排期
	v.SetDefault("schedule.enabled", false)
	v.SetDefault("schedule.interval", 30)
}

// readEnvOverrides 读取环境变量覆盖
func readEnvOverrides(v *viper.Viper) {
	// 支持直接使用环境变量覆盖配置
	// 格式: CONTFUL_DATABASE_HOST, CONTFUL_SECRET 等
	envMappings := map[string]string{
		"DB_HOST":          "database.host",
		"DB_PORT":          "database.port",
		"DB_USER":          "database.user",
		"DB_PASSWORD":      "database.password",
		"DB_NAME":          "database.name",
		"DB_SSL_MODE":      "database.ssl_mode",
		"REDIS_HOST":       "redis.host",
		"REDIS_PORT":       "redis.port",
		"REDIS_PASSWORD":   "redis.password",
		"REDIS_DB":         "redis.db",
		"SERVER_PORT":      "server.port",
		"SECRET":           "security.secret",
		"SECRET_ALGORITHM": "security.algorithm",
		"CRYPTO_MODE":      "security.crypto_mode",
		// RSA 密钥路径
		"RSA_PUBKEY_PATH":  "security.rsa_pubkey_path",
		"RSA_PRIVKEY_PATH": "security.rsa_privkey_path",
		// 存储配置
		"STORAGE_DRIVER":             "storage.driver",
		"STORAGE_UPLOAD_DIR":         "storage.upload_dir",
		"STORAGE_MAX_UPLOAD_SIZE_MB": "storage.max_upload_size_mb",
		"STORAGE_BASE_URL":           "storage.base_url",
	}

	for envKey, configKey := range envMappings {
		if val := os.Getenv(envKey); val != "" {
			v.Set(configKey, val)
		}
	}
}

// 辅助函数

// deriveKey 使用 HKDF-SHA256 从主密钥派生指定用途的密钥
func deriveKey(master, info string, length int) string {
	// 使用固定 salt
	salt := []byte("contful-kdf-salt-v1")

	// HKDF-Extract
	prk := hkdfExpand(sha256.New, []byte(master), salt, nil, 32)

	// HKDF-Expand
	h := hkdfExpand(sha256.New, prk, nil, []byte(info), length)

	result := make([]byte, length)
	copy(result, h)
	return hex.EncodeToString(result)
}

// hkdfExpand 简化的 HKDF-Expand 实现
func hkdfExpand(prf func() hash.Hash, ikm []byte, salt, info []byte, length int) []byte {
	h := prf()
	h.Write(salt)
	h.Write(ikm)
	prk := h.Sum(nil)

	h = prf()
	h.Write(prk)
	h.Write([]byte{1})
	if info != nil {
		h.Write(info)
	}
	return h.Sum(nil)[:length]
}

func getExt(filename string) string {
	i := strings.LastIndex(filename, ".")
	if i == -1 {
		return ""
	}
	return filename[i:]
}
