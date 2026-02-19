package config

import "github.com/spf13/viper"

type Config struct {
	Port          string
	MongoAtlasURI string
	ServiceName   string
	LogLevel      string
}

func Load() *Config {
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	return &Config{
		Port:          viper.GetString("PORT"),
		MongoAtlasURI: viper.GetString("MONGODB_ATLAS_URI"),
		ServiceName:   viper.GetString("SERVICE_NAME"),
		LogLevel:      viper.GetString("LOG_LEVEL"),
	}
}
