package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Storage  StorageConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type StorageConfig struct {
	UploadDir     string
	MaxUploadSize int64
}

func Load() *Config {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./..")
	viper.AddConfigPath("../..")
	viper.AddConfigPath("../../..")
	viper.AutomaticEnv()

	// 默认值
	viper.SetDefault("OPEN_API_PORT", "8081")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("UPLOAD_DIR", "./uploads")
	viper.SetDefault("MAX_UPLOAD_SIZE_MB", 10)

	var cfg Config
	if err := viper.ReadInConfig(); err != nil {
		// .env 文件不存在时使用默认值
	}

	cfg.Server.Port = viper.GetString("OPEN_API_PORT")
	cfg.Database.Host = viper.GetString("DB_HOST")
	cfg.Database.Port = viper.GetInt("DB_PORT")
	cfg.Database.User = viper.GetString("DB_USER")
	cfg.Database.Password = viper.GetString("DB_PASSWORD")
	cfg.Database.Name = viper.GetString("DB_NAME")
	cfg.Database.SSLMode = viper.GetString("DB_SSL_MODE")
	cfg.Redis.Host = viper.GetString("REDIS_HOST")
	cfg.Redis.Port = viper.GetInt("REDIS_PORT")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")
	cfg.Storage.UploadDir = viper.GetString("UPLOAD_DIR")
	cfg.Storage.MaxUploadSize = viper.GetInt64("MAX_UPLOAD_SIZE_MB") * 1024 * 1024

	return &cfg
}
